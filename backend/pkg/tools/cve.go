package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	cveAPIURL         = "https://services.nvd.nist.gov/rest/json/cves/2.0"
	cveRequestTimeout = 20 * time.Second
	maxCVELookup      = 10
)

// cve represents the NVD CVE detail lookup tool
type cve struct {
	cfg       *config.Config
	flowID    int64
	taskID    *int64
	subtaskID *int64
	slp       SearchLogProvider
}

// NewCVETool creates a new CVE detail lookup tool instance
func NewCVETool(
	cfg *config.Config,
	flowID int64,
	taskID, subtaskID *int64,
	slp SearchLogProvider,
) Tool {
	return &cve{
		cfg:       cfg,
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		slp:       slp,
	}
}

// Handle processes a CVE detail lookup request from an AI agent
func (c *cve) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !c.IsAvailable() {
		return "", fmt.Errorf("CVE lookup is not available")
	}

	var action CveAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(c.flowID, c.taskID, c.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal CVE lookup action")
		return "", fmt.Errorf("failed to unmarshal %s lookup action arguments: %w", name, err)
	}

	rawCVEs := strings.ToUpper(strings.TrimSpace(string(action.CVEs)))
	if rawCVEs == "" {
		return "Error: no CVE identifiers provided. Provide one or more CVE IDs (e.g. CVE-2021-44228).", nil
	}

	replacer := strings.NewReplacer(",", " ", ";", " ")
	cveList := strings.Fields(replacer.Replace(rawCVEs))
	if len(cveList) > maxCVELookup {
		cveList = cveList[:maxCVELookup]
	}

	logger = logger.WithFields(logrus.Fields{
		"cves": cveList,
	})

	result, err := c.lookup(ctx, cveList)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("CVE lookup error"),
			langfuse.WithEventInput(strings.Join(cveList, ",")),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name": CveToolName,
				"cves":      cveList,
				"error":     err.Error(),
			}),
		)
		logger.WithError(err).Error("failed to lookup CVE details")
		return fmt.Sprintf("failed to lookup CVE details: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = c.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeCve,
			strings.Join(cveList, ","),
			result,
			c.taskID,
			c.subtaskID,
		)
	}

	return result, nil
}

// lookup queries the NVD CVE API for each CVE id and returns a formatted markdown result
func (c *cve) lookup(ctx context.Context, cveList []string) (string, error) {
	client, err := system.GetHTTPClient(c.cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create http client: %w", err)
	}
	client.Timeout = cveRequestTimeout

	var sections []string
	for _, cveID := range cveList {
		section, err := c.lookupOne(ctx, client, cveID)
		if err != nil {
			sections = append(sections, fmt.Sprintf("### %s\n\n> %v\n", cveID, err))
			continue
		}
		sections = append(sections, section)
	}

	return strings.Join(sections, "\n"), nil
}

func (c *cve) lookupOne(ctx context.Context, client *http.Client, cveID string) (string, error) {
	query := url.Values{}
	query.Set("cveId", cveID)
	apiURL := cveAPIURL + "?" + query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PentAGI/1.0 (CVE lookup)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to NVD API failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("CVE %s not found in NVD", cveID)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("NVD API returned HTTP %d for %s", resp.StatusCode, cveID)
	}

	var apiResp cveResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode NVD response: %w", err)
	}

	if len(apiResp.Vulnerabilities) == 0 {
		return "", fmt.Errorf("CVE %s not found in NVD", cveID)
	}

	return formatCVEDetails(apiResp.Vulnerabilities[0].Cve), nil
}

// IsAvailable returns true if the CVE tool is enabled
func (c *cve) IsAvailable() bool {
	return c.cfg != nil && c.cfg.CveEnabled
}

// cveResponse mirrors the NVD CVE API 2.0 response envelope
type cveResponse struct {
	Vulnerabilities []struct {
		Cve cveDetail `json:"cve"`
	} `json:"vulnerabilities"`
}

type cveDescription struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

type cveReference struct {
	URL string `json:"url"`
}

type cveWeakness struct {
	Description []cveDescription `json:"description"`
}

type cvssV31Data struct {
	Version               string  `json:"version"`
	VectorString          string  `json:"vectorString"`
	BaseScore             float64 `json:"baseScore"`
	BaseSeverity          string  `json:"baseSeverity"`
	AttackVector          string  `json:"attackVector"`
	AttackComplexity      string  `json:"attackComplexity"`
	PrivilegesRequired    string  `json:"privilegesRequired"`
	UserInteraction       string  `json:"userInteraction"`
	Scope                 string  `json:"scope"`
	ConfidentialityImpact string  `json:"confidentialityImpact"`
	IntegrityImpact       string  `json:"integrityImpact"`
	AvailabilityImpact    string  `json:"availabilityImpact"`
}

type cvssMetricV31 struct {
	Source   string     `json:"source"`
	Type     string     `json:"type"`
	CvssData cvssV31Data `json:"cvssData"`
}

type cvssV30Data struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
	BaseSeverity string  `json:"baseSeverity"`
}

type cvssMetricV30 struct {
	Source   string     `json:"source"`
	Type     string     `json:"type"`
	CvssData cvssV30Data `json:"cvssData"`
}

type cvssV2Data struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
	BaseSeverity string  `json:"baseSeverity"`
}

type cvssMetricV2 struct {
	Source   string     `json:"source"`
	Type     string     `json:"type"`
	CvssData cvssV2Data `json:"cvssData"`
}

type cveDetail struct {
	ID           string            `json:"id"`
	Published    string            `json:"published"`
	LastModified string            `json:"lastModified"`
	Descriptions []cveDescription  `json:"descriptions"`
	Metrics      struct {
		CvssMetricV31 []cvssMetricV31 `json:"cvssMetricV31"`
		CvssMetricV30 []cvssMetricV30 `json:"cvssMetricV30"`
		CvssMetricV2  []cvssMetricV2  `json:"cvssMetricV2"`
	} `json:"metrics"`
	References []cveReference `json:"references"`
	Weaknesses []cveWeakness  `json:"weaknesses"`
}

func formatCVEDetails(d cveDetail) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### %s\n", d.ID))

	desc := pickEnglishValue(d.Descriptions)
	if desc != "" {
		sb.WriteString("\n" + desc + "\n")
	}

	sb.WriteString("\n| Field | Value |\n| --- | --- |\n")

	if len(d.Metrics.CvssMetricV31) > 0 {
		m := d.Metrics.CvssMetricV31[0]
		sb.WriteString(fmt.Sprintf("| CVSS v%s | **%.1f** (%s) |\n", m.CvssData.Version, m.CvssData.BaseScore, m.CvssData.BaseSeverity))
		sb.WriteString(fmt.Sprintf("| Vector | `%s` |\n", m.CvssData.VectorString))
		sb.WriteString(fmt.Sprintf("| Attack vector | %s |\n", titleCase(m.CvssData.AttackVector)))
		sb.WriteString(fmt.Sprintf("| Attack complexity | %s |\n", titleCase(m.CvssData.AttackComplexity)))
		sb.WriteString(fmt.Sprintf("| Privileges required | %s |\n", titleCase(m.CvssData.PrivilegesRequired)))
		sb.WriteString(fmt.Sprintf("| User interaction | %s |\n", titleCase(m.CvssData.UserInteraction)))
		sb.WriteString(fmt.Sprintf("| Scope | %s |\n", m.CvssData.Scope))
		sb.WriteString(fmt.Sprintf("| Confidentiality / Integrity / Availability | %s / %s / %s |\n",
			titleCase(m.CvssData.ConfidentialityImpact), titleCase(m.CvssData.IntegrityImpact), titleCase(m.CvssData.AvailabilityImpact)))
	} else if len(d.Metrics.CvssMetricV30) > 0 {
		m := d.Metrics.CvssMetricV30[0]
		sb.WriteString(fmt.Sprintf("| CVSS v%s | **%.1f** (%s) |\n", m.CvssData.Version, m.CvssData.BaseScore, m.CvssData.BaseSeverity))
		sb.WriteString(fmt.Sprintf("| Vector | `%s` |\n", m.CvssData.VectorString))
	} else if len(d.Metrics.CvssMetricV2) > 0 {
		m := d.Metrics.CvssMetricV2[0]
		sb.WriteString(fmt.Sprintf("| CVSS v%s | **%.1f** (%s) |\n", m.CvssData.Version, m.CvssData.BaseScore, m.CvssData.BaseSeverity))
		sb.WriteString(fmt.Sprintf("| Vector | `%s` |\n", m.CvssData.VectorString))
	}

	if d.Published != "" {
		sb.WriteString(fmt.Sprintf("| Published | %s |\n", d.Published))
	}
	if d.LastModified != "" {
		sb.WriteString(fmt.Sprintf("| Last modified | %s |\n", d.LastModified))
	}

	if len(d.Weaknesses) > 0 {
		var cwes []string
		for _, w := range d.Weaknesses {
			for _, desc := range w.Description {
				if desc.Lang == "en" && desc.Value != "" {
					cwes = append(cwes, desc.Value)
				}
			}
		}
		if len(cwes) > 0 {
			sb.WriteString("| CWE | " + strings.Join(dedupe(cwes), ", ") + " |\n")
		}
	}

	if len(d.References) > 0 {
		var urls []string
		for _, ref := range d.References {
			if ref.URL != "" {
				urls = append(urls, ref.URL)
			}
		}
		if len(urls) > 0 {
			sb.WriteString("\n**References:**\n")
			for _, u := range urls {
				sb.WriteString("- " + u + "\n")
			}
		}
	}

	return sb.String()
}

func pickEnglishValue(items []cveDescription) string {
	for _, it := range items {
		if it.Lang == "en" && it.Value != "" {
			return it.Value
		}
	}
	for _, it := range items {
		if it.Value != "" {
			return it.Value
		}
	}
	return ""
}

func titleCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func dedupe(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, it := range items {
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	sort.Strings(out)
	return out
}
