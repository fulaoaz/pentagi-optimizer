package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	epssAPIURL         = "https://api.first.org/data/v1/epss"
	epssRequestTimeout = 15 * time.Second
	defaultEPSSLimit   = 10
	maxEPSSLimit       = 50
)

// epss represents the EPSS (Exploit Prediction Scoring System) lookup tool
type epss struct {
	cfg       *config.Config
	flowID    int64
	taskID    *int64
	subtaskID *int64
	slp       SearchLogProvider
}

// NewEPPSTool creates a new EPSS lookup tool instance
func NewEPPSTool(
	cfg *config.Config,
	flowID int64,
	taskID, subtaskID *int64,
	slp SearchLogProvider,
) Tool {
	return &epss{
		cfg:       cfg,
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		slp:       slp,
	}
}

// Handle processes an EPSS score lookup request from an AI agent
func (e *epss) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	if !e.IsAvailable() {
		return "", fmt.Errorf("EPSS is not available")
	}

	var action EppssAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(e.flowID, e.taskID, e.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal EPSS lookup action")
		return "", fmt.Errorf("failed to unmarshal %s lookup action arguments: %w", name, err)
	}

	// Normalise CVE list: uppercase, trim spaces, split on comma/space
	rawCVEs := strings.ToUpper(strings.TrimSpace(action.CVEs))
	if rawCVEs == "" {
		return "Error: no CVE identifiers provided. Provide one or more CVE IDs (e.g. CVE-2021-44228).", nil
	}

	replacer := strings.NewReplacer(",", " ", ";", " ")
	cveList := strings.Fields(replacer.Replace(rawCVEs))
	if len(cveList) > maxEPSSLimit {
		cveList = cveList[:maxEPSSLimit]
	}

	logger = logger.WithFields(logrus.Fields{
		"cves": cveList,
	})

	result, err := e.lookup(ctx, cveList)
	if err != nil {
		observation.Event(
			langfuse.WithEventName("EPSS lookup error"),
			langfuse.WithEventInput(strings.Join(cveList, ",")),
			langfuse.WithEventStatus(err.Error()),
			langfuse.WithEventLevel(langfuse.ObservationLevelWarning),
			langfuse.WithEventMetadata(langfuse.Metadata{
				"tool_name": EppssToolName,
				"cves":      cveList,
				"error":     err.Error(),
			}),
		)
		logger.WithError(err).Error("failed to lookup EPSS scores")
		return fmt.Sprintf("failed to lookup EPSS scores: %v", err), nil
	}

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = e.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeEppss,
			strings.Join(cveList, ","),
			result,
			e.taskID,
			e.subtaskID,
		)
	}

	return result, nil
}

// lookup calls the FIRST.org EPSS API and returns a formatted markdown result
func (e *epss) lookup(ctx context.Context, cveList []string) (string, error) {
	query := url.Values{}
	query.Set("cve", strings.Join(cveList, ","))
	apiURL := epssAPIURL + "?" + query.Encode()

	client, err := system.GetHTTPClient(e.cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create http client: %w", err)
	}
	client.Timeout = epssRequestTimeout

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PentAGI/1.0 (EPSS lookup)")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to EPSS API failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("EPSS API returned HTTP %d", resp.StatusCode)
	}

	var apiResp epssResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode EPSS response: %w", err)
	}

	return formatEPSSResults(cveList, apiResp), nil
}

// IsAvailable returns true if the EPSS tool is enabled
func (e *epss) IsAvailable() bool {
	return e.enabled()
}

func (e *epss) enabled() bool {
	return e.cfg != nil && e.cfg.EppssEnabled
}

// epssScore represents a single EPSS score record
// The API returns epss and percentile as strings in scientific notation
type epssScore struct {
	CVE        string `json:"cve"`
	EPSS       string `json:"epss"`
	Percentile string `json:"percentile"`
	Date       string `json:"date"`
}

// epssResponse is the top-level JSON response from the EPSS API
type epssResponse struct {
	Status   string      `json:"status"`
	Total    int         `json:"total"`
	Data     []epssScore `json:"data"`
	NotFound []string    `json:"not_found,omitempty"`
}

// formatEPSSResults converts an epssResponse into a human-readable markdown string
func formatEPSSResults(cveList []string, resp epssResponse) string {
	var sb strings.Builder

	sb.WriteString("# EPSS (Exploit Prediction Scoring System) Scores\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** FIRST.org EPSS API  \n"))
	sb.WriteString(fmt.Sprintf("**Date:** %s  \n", time.Now().Format("2006-01-02")))
	sb.WriteString("\n**EPSS** measures the probability that a vulnerability will be exploited in the wild (0.0 to 1.0). ")
	sb.WriteString("**Percentile** ranks a CVE against all published CVEs (0.0 to 1.0). ")
	sb.WriteString("Higher scores mean higher exploitation likelihood — prioritize accordingly.\n\n")
	sb.WriteString("| CVE | EPSS Score | Percentile | Data Date | Exploitation Risk |\n")
	sb.WriteString("|-----|------------|------------|-----------|-------------------|\n")

	for _, score := range resp.Data {
		epssVal := parseEPSSFloat(score.EPSS)
		percentileVal := parseEPSSFloat(score.Percentile)
		sb.WriteString(fmt.Sprintf("| %s | %.4f | %.4f | %s | %s |\n",
			score.CVE, epssVal, percentileVal, score.Date, epssRiskLabel(epssVal)))
	}

	if len(resp.NotFound) > 0 {
		sb.WriteString("\n**Not found in EPSS database:** " + strings.Join(resp.NotFound, ", ") + "\n")
	}

	return sb.String()
}

// parseEPSSFloat parses EPSS API values (which may be in scientific notation) into a float
func parseEPSSFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// epssRiskLabel maps an EPSS score to a human-readable risk level
func epssRiskLabel(score float64) string {
	switch {
	case score >= 0.9:
		return "Critical"
	case score >= 0.7:
		return "High"
	case score >= 0.3:
		return "Medium"
	case score > 0:
		return "Low"
	default:
		return "Unknown"
	}
}
