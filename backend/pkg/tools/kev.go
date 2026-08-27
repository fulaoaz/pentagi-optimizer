package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"
	"pentagi/pkg/system"

	"github.com/sirupsen/logrus"
)

const (
	kevAPIURL         = "https://www.cisa.gov/sites/default/files/feeds/known_exploited_vulnerabilities.json"
	kevRequestTimeout = 30 * time.Second
	defaultKEVLimit   = 10
	maxKEVLimit       = 50
)

// kev represents the CISA Known Exploited Vulnerabilities catalog lookup tool
type kev struct {
	cfg       *config.Config
	flowID    int64
	taskID    *int64
	subtaskID *int64
	slp       SearchLogProvider
}

// NewKEVTool creates a new KEV catalog lookup tool instance
func NewKEVTool(
	cfg *config.Config,
	flowID int64,
	taskID, subtaskID *int64,
	slp SearchLogProvider,
) Tool {
	return &kev{
		cfg:       cfg,
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		slp:       slp,
	}
}

// Handle processes a KEV catalog lookup request from an AI agent
func (k *kev) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !k.IsAvailable() {
		return "", fmt.Errorf("KEV lookup is not available")
	}

	var action KevAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(k.flowID, k.taskID, k.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal KEV lookup action")
		return "", fmt.Errorf("failed to unmarshal %s lookup action arguments: %w", name, err)
	}

	limit := int(action.Limit)
	if limit <= 0 {
		limit = defaultKEVLimit
	}
	if limit > maxKEVLimit {
		limit = maxKEVLimit
	}

	logger = logger.WithFields(logrus.Fields{
		"cves":   action.CVEs,
		"vendor": action.Vendor,
		"limit":  limit,
	})

	result, err := k.lookup(ctx, string(action.CVEs), string(action.Vendor), string(action.Product), limit)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("KEV lookup error"),
			langfuse.WithEventInput(fmt.Sprintf("cves=%s vendor=%s", action.CVEs, action.Vendor)),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name": KevToolName,
				"cves":      action.CVEs,
				"vendor":    action.Vendor,
				"error":     err.Error(),
			}),
		)
		logger.WithError(err).Error("failed to lookup KEV catalog")
		return fmt.Sprintf("failed to lookup KEV catalog: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = k.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeKev,
			fmt.Sprintf("cves=%s vendor=%s", action.CVEs, action.Vendor),
			result,
			k.taskID,
			k.subtaskID,
		)
	}

	return result, nil
}

// lookup fetches the full KEV catalog and filters it by the given criteria
func (k *kev) lookup(ctx context.Context, cves, vendor, product string, limit int) (string, error) {
	client, err := system.GetHTTPClient(k.cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create http client: %w", err)
	}
	client.Timeout = kevRequestTimeout

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, kevAPIURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PentAGI/1.0 (KEV lookup)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to CISA KEV feed failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CISA KEV feed returned HTTP %d", resp.StatusCode)
	}

	var apiResp kevResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode CISA KEV response: %w", err)
	}

	return formatKEVResults(apiResp.Vulnerabilities, cves, vendor, product, limit), nil
}

// IsAvailable returns true if the KEV tool is enabled
func (k *kev) IsAvailable() bool {
	return k.cfg != nil && k.cfg.KevEnabled
}

// kevResponse mirrors the CISA KEV catalog JSON structure
type kevResponse struct {
	Title           string `json:"title"`
	CatalogVersion  string `json:"catalogVersion"`
	DateReleased    string `json:"dateReleased"`
	Vulnerabilities []struct {
		CveID                      string `json:"cveID"`
		VendorProject              string `json:"vendorProject"`
		Product                    string `json:"product"`
		VulnerabilityName          string `json:"vulnerabilityName"`
		DateAdded                  string `json:"dateAdded"`
		ShortDescription           string `json:"shortDescription"`
		RequiredAction             string `json:"requiredAction"`
		DueDate                    string `json:"dueDate"`
		KnownRansomwareCampaignUse string `json:"knownRansomwareCampaignUse"`
		Notes                      string `json:"notes"`
	} `json:"vulnerabilities"`
}

func formatKEVResults(vulns []struct {
	CveID                      string `json:"cveID"`
	VendorProject              string `json:"vendorProject"`
	Product                    string `json:"product"`
	VulnerabilityName          string `json:"vulnerabilityName"`
	DateAdded                  string `json:"dateAdded"`
	ShortDescription           string `json:"shortDescription"`
	RequiredAction             string `json:"requiredAction"`
	DueDate                    string `json:"dueDate"`
	KnownRansomwareCampaignUse string `json:"knownRansomwareCampaignUse"`
	Notes                      string `json:"notes"`
}, cves, vendor, product string, limit int) string {
	cveFilter := parseCVEList(cves)
	vendorFilter := strings.ToLower(strings.TrimSpace(vendor))
	productFilter := strings.ToLower(strings.TrimSpace(product))

	var matches []struct {
		CveID                      string `json:"cveID"`
		VendorProject              string `json:"vendorProject"`
		Product                    string `json:"product"`
		VulnerabilityName          string `json:"vulnerabilityName"`
		DateAdded                  string `json:"dateAdded"`
		ShortDescription           string `json:"shortDescription"`
		RequiredAction             string `json:"requiredAction"`
		DueDate                    string `json:"dueDate"`
		KnownRansomwareCampaignUse string `json:"knownRansomwareCampaignUse"`
		Notes                      string `json:"notes"`
	}

	for _, v := range vulns {
		if len(cveFilter) > 0 && !cveFilter[v.CveID] {
			continue
		}
		if vendorFilter != "" && !strings.Contains(strings.ToLower(v.VendorProject), vendorFilter) {
			continue
		}
		if productFilter != "" && !strings.Contains(strings.ToLower(v.Product), productFilter) {
			continue
		}
		matches = append(matches, v)
	}

	// Sort by DateAdded descending (newest first)
	sort.SliceStable(matches, func(i, j int) bool {
		return matches[i].DateAdded > matches[j].DateAdded
	})

	if len(matches) > limit {
		matches = matches[:limit]
	}

	if len(matches) == 0 {
		return "No known exploited vulnerabilities matched the given criteria in the CISA KEV catalog."
	}

	var sb strings.Builder
	sb.WriteString("### CISA Known Exploited Vulnerabilities (KEV)\n\n")
	sb.WriteString("These vulnerabilities are confirmed to be exploited in the wild; treat them as highest priority when found.\n\n")
	sb.WriteString("| CVE | Vendor / Product | Name | Added | Due | Ransomware |\n")
	sb.WriteString("| --- | --- | --- | --- | --- | --- |\n")
	for _, v := range matches {
		ransom := "no"
		if strings.EqualFold(v.KnownRansomwareCampaignUse, "Known") {
			ransom = "yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s / %s | %s | %s | %s | %s |\n",
			v.CveID, v.VendorProject, v.Product, v.VulnerabilityName, v.DateAdded, v.DueDate, ransom))
	}

	// Append detail sections for the first few entries so the agent has actionable context
	maxDetail := 5
	if len(matches) < maxDetail {
		maxDetail = len(matches)
	}
	for _, v := range matches[:maxDetail] {
		sb.WriteString(fmt.Sprintf("\n**%s** — %s / %s\n", v.CveID, v.VendorProject, v.Product))
		if v.ShortDescription != "" {
			sb.WriteString(v.ShortDescription + "\n")
		}
		if v.RequiredAction != "" {
			sb.WriteString("Required action: " + v.RequiredAction + "\n")
		}
		if v.Notes != "" {
			sb.WriteString("Notes: " + v.Notes + "\n")
		}
	}

	return sb.String()
}

func parseCVEList(s string) map[string]bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return nil
	}
	out := make(map[string]bool)
	replacer := strings.NewReplacer(",", " ", ";", " ")
	for _, c := range strings.Fields(replacer.Replace(s)) {
		out[c] = true
	}
	return out
}
