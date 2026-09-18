package tools

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
)

const (
	minMdContentSize         = 50
	minHtmlContentSize       = 300
	minImgContentSize        = 2048
	maxScraperErrorBodyBytes = 512
	maxScraperResponseBytes  = 16 * 1024 * 1024
	maxScraperHeaderBytes    = 64 * 1024
	maxBrowserTargetURLBytes = 8 * 1024
)

// nonHTMLExtensions lists URL path suffixes that point to resources the scraper
// cannot render as text. When a URL matches, the browser tool returns a
// descriptive hint instead of a generic "content too small" error.
var nonHTMLExtensions = []string{
	".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
	".zip", ".tar", ".gz", ".bz2", ".rar", ".7z",
	".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".svg", ".ico",
	".mp3", ".mp4", ".avi", ".mov", ".mkv", ".wav",
	".exe", ".bin", ".dll", ".so", ".dmg", ".apk",
}

// isBinaryURL returns true when the URL points to a known non-HTML resource.
func isBinaryURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	if idx := strings.Index(lower, "?"); idx != -1 {
		lower = lower[:idx]
	}
	for _, ext := range nonHTMLExtensions {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

var localZones = []string{
	".localdomain",
	".local",
	".lan",
	".htb",
	".dev",
	".test",
	".corp",
	".example",
	".invalid",
	".internal",
	".home.arpa",
}

var nonPublicTargetCIDRs = []string{
	"0.0.0.0/8",
	"100.64.0.0/10",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"2001:2::/48",
	"2001:10::/28",
	"2001:db8::/32",
	"ff00::/8",
}

type browser struct {
	flowID    int64
	taskID    *int64
	subtaskID *int64
	dataDir   string
	scPrvURL  string
	scPubURL  string
	scp       ScreenshotProvider
}

func NewBrowserTool(
	flowID int64, taskID, subtaskID *int64,
	dataDir, scPrvURL, scPubURL string,
	scp ScreenshotProvider,
) Tool {
	return &browser{
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		dataDir:   dataDir,
		scPrvURL:  scPrvURL,
		scPubURL:  scPubURL,
		scp:       scp,
	}
}

func (b *browser) wrapCommandResult(ctx context.Context, name, result, url, screen string, err error) (string, error) {
	ctx, observation := obs.Observer.NewObservation(ctx)
	redactedURL := redactURL(url)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("browser tool error swallowed"),
			langfuse.WithEventInput(map[string]any{
				"url":    redactedURL,
				"action": name,
			}),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name": BrowserToolName,
				"url":       redactedURL,
				"screen":    screen,
				"error":     err.Error(),
			}),
		)

		logrus.WithContext(ctx).WithError(err).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
			"tool":   name,
			"url":    redactedURL,
			"screen": screen,
			"result": result[:min(len(result), 1000)],
		})).Error("browser tool failed")
		return fmt.Sprintf("browser tool '%s' handled with error: %v", name, err), nil
	}
	if screen != "" {
		_, _ = b.scp.PutScreenshot(ctx, screen, redactedURL, b.taskID, b.subtaskID)
	}
	return result, nil
}

func (b *browser) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !b.IsAvailable() {
		return "", fmt.Errorf("browser is not available")
	}

	var action Browser
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool": name,
	}))

	if name != "browser" {
		logger.Error("unknown tool")
		return "", fmt.Errorf("unknown tool: %s", name)
	}

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal browser action")
		return "", fmt.Errorf("failed to unmarshal browser action: %w", err)
	}

	// LLMs occasionally wrap the URL with an incidental leading/trailing
	// newline, carriage return, or space (e.g. copy-pasted from command
	// output); net/url.Parse rejects control characters outright, failing the
	// whole call over whitespace that carries no meaning. Strip it here so it
	// never reaches resolveUrl.
	action.Url = strings.TrimSpace(action.Url)

	if action.Action == "" {
		// The LLM occasionally omits the required 'action' field even though the
		// tool schema marks it required. 'markdown' is the safest default: it is
		// the most commonly used and most versatile content format, so infer it
		// instead of failing the call outright and burning a tool-call-fixer
		// round-trip on something that doesn't need one.
		action.Action = Markdown
	}

	logger = logger.WithFields(logrus.Fields{
		"action": action.Action,
		"url":    redactURL(action.Url),
	})

	switch action.Action {
	case Markdown:
		result, screen, err := b.ContentMD(ctx, action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	case HTML:
		result, screen, err := b.ContentHTML(ctx, action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	case Links:
		result, screen, err := b.Links(ctx, action.Url)
		return b.wrapCommandResult(ctx, name, result, action.Url, screen, err)
	default:
		logger.Error("unknown browser action")
		return "", fmt.Errorf("unknown browser action: %s", action.Action)
	}
}

func (b *browser) ContentMD(ctx context.Context, url string) (string, string, error) {
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool":   "browser",
		"action": "markdown",
		"url":    redactURL(url),
	}))
	logger.Debug("trying to get markdown content")

	var (
		wg                        sync.WaitGroup
		content, screenshotName   string
		errContent, errScreenshot error
	)
	wg.Add(2)

	go func() {
		defer wg.Done()
		content, errContent = b.getMD(url)
	}()

	go func() {
		defer wg.Done()
		screenshotName, errScreenshot = b.getScreenshot(url)
	}()

	wg.Wait()

	if errContent != nil {
		return "", "", errContent
	}
	if errScreenshot != nil {
		logger.WithError(errScreenshot).Warn("failed to capture screenshot, continuing without it")
		screenshotName = ""
	}

	return content, screenshotName, nil
}

func (b *browser) ContentHTML(ctx context.Context, url string) (string, string, error) {
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool":   "browser",
		"action": "html",
		"url":    redactURL(url),
	}))
	logger.Debug("trying to get HTML content")

	var (
		wg                        sync.WaitGroup
		content, screenshotName   string
		errContent, errScreenshot error
	)
	wg.Add(2)

	go func() {
		defer wg.Done()
		content, errContent = b.getHTML(url)
	}()

	go func() {
		defer wg.Done()
		screenshotName, errScreenshot = b.getScreenshot(url)
	}()

	wg.Wait()

	if errContent != nil {
		return "", "", errContent
	}
	if errScreenshot != nil {
		logger.WithError(errScreenshot).Warn("failed to capture screenshot, continuing without it")
		screenshotName = ""
	}

	return content, screenshotName, nil
}

func (b *browser) Links(ctx context.Context, url string) (string, string, error) {
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(b.flowID, b.taskID, b.subtaskID, logrus.Fields{
		"tool":   "browser",
		"action": "links",
		"url":    redactURL(url),
	}))
	logger.Debug("trying to get links")

	var (
		wg                      sync.WaitGroup
		links, screenshotName   string
		errLinks, errScreenshot error
	)
	wg.Add(2)

	go func() {
		defer wg.Done()
		links, errLinks = b.getLinks(url)
	}()

	go func() {
		defer wg.Done()
		screenshotName, errScreenshot = b.getScreenshot(url)
	}()

	wg.Wait()

	if errLinks != nil {
		return "", "", errLinks
	}
	if errScreenshot != nil {
		logger.WithError(errScreenshot).Warn("failed to capture screenshot, continuing without it")
		screenshotName = ""
	}

	return links, screenshotName, nil
}

func (b *browser) resolveUrl(targetURL string) (*url.URL, error) {
	u, err := parseBrowserTargetURL(targetURL)
	if err != nil {
		return nil, err
	}

	host := normalizeBrowserHost(u.Hostname())
	isPrivate := browserTargetHostIsPrivate(host)

	var scraperURL string
	if isPrivate {
		scraperURL = b.scPrvURL
		if scraperURL == "" {
			scraperURL = b.scPubURL
		}
	} else {
		scraperURL = b.scPubURL
		if scraperURL == "" {
			scraperURL = b.scPrvURL
		}
	}

	if scraperURL == "" {
		return nil, fmt.Errorf("no scraper URL configured")
	}

	parsedScraperURL, err := parseScraperURL(scraperURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scraper URL: %w", err)
	}
	return parsedScraperURL, nil
}

func parseBrowserTargetURL(rawURL string) (*url.URL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("target URL is empty")
	}
	if len(rawURL) > maxBrowserTargetURLBytes {
		return nil, fmt.Errorf("target URL exceeds the %d-byte limit", maxBrowserTargetURLBytes)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %w", err)
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return nil, fmt.Errorf("target URL must use http or https")
	}
	if u.Host == "" || u.Hostname() == "" {
		return nil, fmt.Errorf("target URL must include a host")
	}
	if err := validateBrowserHost(u.Hostname()); err != nil {
		return nil, fmt.Errorf("invalid target URL host: %w", err)
	}
	if u.User != nil {
		return nil, fmt.Errorf("target URL must not include user information")
	}
	if strings.ContainsAny(u.Hostname(), " \t\r\n") {
		return nil, fmt.Errorf("target URL host contains whitespace")
	}
	if port := u.Port(); port != "" {
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return nil, fmt.Errorf("target URL contains an invalid port")
		}
	}

	return u, nil
}

func parseScraperURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return nil, fmt.Errorf("scraper URL must use http or https")
	}
	if u.Host == "" || u.Hostname() == "" {
		return nil, fmt.Errorf("scraper URL must include a host")
	}
	if err := validateBrowserHost(u.Hostname()); err != nil {
		return nil, fmt.Errorf("invalid scraper URL host: %w", err)
	}
	if strings.ContainsAny(u.Hostname(), " \t\r\n") {
		return nil, fmt.Errorf("scraper URL host contains whitespace")
	}
	if port := u.Port(); port != "" {
		portNumber, err := strconv.Atoi(port)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return nil, fmt.Errorf("scraper URL contains an invalid port")
		}
	}
	return u, nil
}

func normalizeBrowserHost(host string) string {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if zoneIndex := strings.LastIndex(host, "%"); zoneIndex >= 0 {
		host = host[:zoneIndex]
	}
	if asciiHost, err := idna.Lookup.ToASCII(host); err == nil {
		host = asciiHost
	}
	return host
}

func validateBrowserHost(rawHost string) error {
	host := strings.TrimSuffix(strings.TrimSpace(rawHost), ".")
	if host == "" {
		return fmt.Errorf("host is empty")
	}
	if strings.ContainsAny(host, " \t\r\n") {
		return fmt.Errorf("host contains whitespace")
	}

	if zoneIndex := strings.LastIndex(host, "%"); zoneIndex >= 0 {
		ipHost, zone := host[:zoneIndex], host[zoneIndex+1:]
		if net.ParseIP(ipHost) == nil || zone == "" || strings.ContainsAny(zone, "%/?#[]") {
			return fmt.Errorf("invalid IPv6 zone")
		}
		return nil
	}
	if net.ParseIP(host) != nil {
		return nil
	}

	asciiHost, err := idna.Lookup.ToASCII(host)
	if err != nil {
		return fmt.Errorf("invalid hostname")
	}
	if len(asciiHost) > 253 {
		return fmt.Errorf("hostname is too long")
	}
	for _, label := range strings.Split(asciiHost, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("invalid hostname label")
		}
		for _, char := range label {
			if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') && (char < '0' || char > '9') && char != '-' {
				return fmt.Errorf("invalid hostname character")
			}
		}
	}
	return nil
}

func browserTargetHostIsPrivate(host string) bool {
	host = normalizeBrowserHost(host)
	if host == "" {
		return true
	}

	if hostIP := net.ParseIP(host); hostIP != nil {
		return isNonPublicTargetIP(hostIP)
	}
	if isKnownLocalBrowserHost(host) {
		return true
	}

	addresses, err := net.LookupIP(host)
	if err == nil && len(addresses) > 0 {
		for _, address := range addresses {
			if isNonPublicTargetIP(address) {
				return true
			}
		}
		return false
	}

	// Preserve the existing public-scraper fallback for dotted names that cannot
	// be resolved locally. The scraper remains responsible for the final fetch.
	return false
}

func isKnownLocalBrowserHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || !strings.Contains(host, ".") {
		return true
	}
	for _, zone := range localZones {
		zone = strings.TrimPrefix(strings.ToLower(zone), ".")
		if host == zone || strings.HasSuffix(host, "."+zone) {
			return true
		}
	}
	return false
}

func isNonPublicTargetIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}
	if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || !ip.IsGlobalUnicast() {
		return true
	}
	for _, cidr := range nonPublicTargetCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "<invalid-url>"
	}
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func (b *browser) saveScreenshotData(screenshot []byte) (string, error) {
	flowDirName := fmt.Sprintf("flow-%d", b.flowID)
	screenshotDir := filepath.Join(b.dataDir, "screenshots", flowDirName)

	err := os.MkdirAll(screenshotDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to prepare screenshot directory: %w", err)
	}

	screenshotName := fmt.Sprintf("screenshot-%d.png", time.Now().Unix())
	path := filepath.Join(screenshotDir, screenshotName)

	err = os.WriteFile(path, screenshot, 0644)
	if err != nil {
		return "", fmt.Errorf("screenshot write operation failed: %w", err)
	}

	return screenshotName, nil
}

func (b *browser) getMD(targetURL string) (string, error) {
	if isBinaryURL(targetURL) {
		return "", fmt.Errorf(
			"the URL appears to point to a binary/non-HTML resource (e.g. PDF, image, archive) " +
				"that cannot be rendered as markdown. Use the terminal tool with curl/wget to download it instead",
		)
	}

	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("url", targetURL)
	scraperURL.Path = "/markdown"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch content by url '%s': %w", redactURL(targetURL), err)
	}
	if len(content) < minMdContentSize {
		return fmt.Sprintf(
			"[WARNING: page returned very little content (%d bytes), it may be a redirect, error page, or near-empty]\n\n%s",
			len(content), string(content),
		), nil
	}

	return string(content), nil
}

func (b *browser) getHTML(targetURL string) (string, error) {
	if isBinaryURL(targetURL) {
		return "", fmt.Errorf(
			"the URL appears to point to a binary/non-HTML resource (e.g. PDF, image, archive) " +
				"that cannot be rendered as HTML. Use the terminal tool with curl/wget to download it instead",
		)
	}

	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("url", targetURL)
	scraperURL.Path = "/html"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch content by url '%s': %w", redactURL(targetURL), err)
	}
	if len(content) < minHtmlContentSize {
		return fmt.Sprintf(
			"[WARNING: page returned very little HTML content (%d bytes), it may be a redirect, error page, or near-empty]\n\n%s",
			len(content), string(content),
		), nil
	}

	return string(content), nil
}

func (b *browser) getLinks(targetURL string) (string, error) {
	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("url", targetURL)
	scraperURL.Path = "/links"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch links by url '%s': %w", redactURL(targetURL), err)
	}

	links := []struct {
		Title string
		Link  string
	}{}
	err = json.Unmarshal(content, &links)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal links: %w", err)
	}

	var buffer strings.Builder
	buffer.WriteString(fmt.Sprintf("Links list from URL '%s'\n", targetURL))
	for _, l := range links {
		link := strings.TrimSpace(l.Link)
		if link == "" {
			continue
		}
		title := strings.TrimSpace(l.Title)
		if title == "" {
			title = "UNTITLED"
		}
		buffer.WriteString(fmt.Sprintf("[%s](%s)\n", title, l.Link))
	}

	return buffer.String(), nil
}

func (b *browser) getScreenshot(targetURL string) (string, error) {
	scraperURL, err := b.resolveUrl(targetURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve url: %w", err)
	}

	query := scraperURL.Query()
	query.Add("fullPage", "true")
	query.Add("url", targetURL)
	scraperURL.Path = "/screenshot"
	scraperURL.RawQuery = query.Encode()

	content, err := b.callScraper(scraperURL.String())
	if err != nil {
		return "", fmt.Errorf("failed to fetch screenshot by url '%s': %w", redactURL(targetURL), err)
	}
	if len(content) < minImgContentSize {
		return "", fmt.Errorf("image size is less than minimum: %d bytes", minImgContentSize)
	}

	return b.saveScreenshotData(content)
}

func (b *browser) callScraper(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 65 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:        &tls.Config{InsecureSkipVerify: true},
			MaxResponseHeaderBytes: maxScraperHeaderBytes,
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data by scraper '%s': %w", redactURL(url), err)
	}
	defer resp.Body.Close()
	redactedURL := redactURL(url)
	if resp.ContentLength > maxScraperResponseBytes {
		return nil, fmt.Errorf("scraper response body exceeds the %d-byte limit for '%s'", maxScraperResponseBytes, redactedURL)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, maxScraperErrorBodyBytes+1))
			if preview := strings.TrimSpace(string(body)); preview != "" {
				if truncated := len(body) > maxScraperErrorBodyBytes; truncated {
					preview = preview[:maxScraperErrorBodyBytes] + "... [truncated]"
				}
				return nil, fmt.Errorf(
					"unexpected resp code for scraper '%s': %d, response: %s", redactedURL, resp.StatusCode, preview,
				)
			}
		}
		return nil, fmt.Errorf("unexpected resp code for scraper '%s': %d", redactedURL, resp.StatusCode)
	}

	content, err := io.ReadAll(io.LimitReader(resp.Body, maxScraperResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for scraper '%s': %w", redactedURL, err)
	} else if len(content) > maxScraperResponseBytes {
		return nil, fmt.Errorf("scraper response body exceeds the %d-byte limit for '%s'", maxScraperResponseBytes, redactedURL)
	} else if len(content) == 0 {
		return nil, fmt.Errorf("empty response body for scraper '%s'", redactedURL)
	}

	return content, nil
}

func (b *browser) IsAvailable() bool {
	return b.scPrvURL != "" || b.scPubURL != ""
}
