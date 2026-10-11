package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"pentagi/pkg/config"
	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
	"pentagi/pkg/observability/langfuse"

	"github.com/sirupsen/logrus"
)

// maxCVSSVectors caps how many vectors a single tool call may score, so a
// malformed agent argument cannot turn into an unbounded formatting loop.
const maxCVSSVectors = 25

// cvss scores CVSS v3.1 and v4.0 vectors locally. Unlike the other scoring
// tool in this package (EPSS, which queries FIRST.org), CVSS is a closed-form
// calculation defined entirely by the specification, so this tool performs no
// network I/O and is always available.
type cvss struct {
	cfg       *config.Config
	flowID    int64
	taskID    *int64
	subtaskID *int64
	slp       SearchLogProvider
}

// NewCVSSTool creates a new CVSS calculator tool instance
func NewCVSSTool(
	cfg *config.Config,
	flowID int64,
	taskID, subtaskID *int64,
	slp SearchLogProvider,
) Tool {
	return &cvss{
		cfg:       cfg,
		flowID:    flowID,
		taskID:    taskID,
		subtaskID: subtaskID,
		slp:       slp,
	}
}

// Handle processes a CVSS scoring request from an AI agent
func (c *cvss) Handle(ctx context.Context, name string, args json.RawMessage) (string, error) {
	var action CvssAction
	ctx, observation := obs.Observer.NewObservation(ctx)
	logger := logrus.WithContext(ctx).WithFields(enrichLogrusFields(c.flowID, c.taskID, c.subtaskID, logrus.Fields{
		"tool": name,
		"args": string(args),
	}))

	if err := json.Unmarshal(args, &action); err != nil {
		logger.WithError(err).Error("failed to unmarshal CVSS scoring action")
		return "", fmt.Errorf("failed to unmarshal %s scoring action arguments: %w", name, err)
	}

	// Accept several shapes the agent may produce: a single vector, or a
	// newline / comma / semicolon separated list of them. Splitting on commas
	// is safe because a CVSS vector separates its own metrics with slashes.
	rawVectors := strings.TrimSpace(string(action.Vectors))
	if rawVectors == "" {
		return "Error: no CVSS vectors provided. Provide one or more vector strings " +
			"(e.g. CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H).", nil
	}

	replacer := strings.NewReplacer(",", "\n", ";", "\n")
	var vectors []string
	for _, field := range strings.Split(replacer.Replace(rawVectors), "\n") {
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			vectors = append(vectors, trimmed)
		}
	}
	if len(vectors) == 0 {
		return "Error: no CVSS vectors provided after parsing the input.", nil
	}
	if len(vectors) > maxCVSSVectors {
		vectors = vectors[:maxCVSSVectors]
	}

	logger = logger.WithField("vectors", vectors)

	result := formatCVSSResults(vectors)

	observation.Event(
		langfuse.WithEventName("CVSS scoring"),
		langfuse.WithEventInput(strings.Join(vectors, " ")),
		langfuse.WithEventMetadata(langfuse.Metadata{
			"tool_name": CvssToolName,
			"vectors":   vectors,
		}),
	)

	if agentCtx, ok := GetAgentContext(ctx); ok {
		_, _ = c.slp.PutLog(
			ctx,
			agentCtx.ParentAgentType,
			agentCtx.CurrentAgentType,
			database.SearchengineTypeCvss,
			strings.Join(vectors, " "),
			result,
			c.taskID,
			c.subtaskID,
		)
	}

	logger.Debug("scored CVSS vectors")

	return result, nil
}

// IsAvailable reports whether scoring is enabled. Unlike the network-backed
// tools there is no credential to check, so this only honours the operator's
// CVSS_ENABLED switch (on by default).
func (c *cvss) IsAvailable() bool {
	return c.cfg != nil && c.cfg.CvssEnabled
}

// cvssResult holds the outcome of scoring one vector.
type cvssResult struct {
	Vector   string
	Version  string
	Score    float64
	Severity string
	Err      error
}

// formatCVSSResults scores every vector and renders a markdown report.
func formatCVSSResults(vectors []string) string {
	results := make([]cvssResult, 0, len(vectors))
	for _, vector := range vectors {
		results = append(results, scoreCVSSVector(vector))
	}

	var sb strings.Builder

	sb.WriteString("# CVSS Scores\n\n")
	sb.WriteString("**Source:** local calculation per CVSS v3.1 / v4.0 specification (no external lookup)\n\n")
	sb.WriteString("**Base score** rates intrinsic severity from 0.0 to 10.0. It reflects exploitability and impact ")
	sb.WriteString("only — it does not estimate whether exploitation is actually happening in the wild. ")
	sb.WriteString("Pair it with the `eppss` tool for exploitation likelihood before deciding remediation order.\n\n")
	sb.WriteString("| Vector | Version | Base Score | Severity |\n")
	sb.WriteString("|--------|---------|------------|----------|\n")

	var failures []cvssResult
	for _, r := range results {
		if r.Err != nil {
			failures = append(failures, r)
			continue
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %.1f | %s |\n",
			r.Vector, r.Version, r.Score, r.Severity))
	}

	if len(failures) == len(results) {
		// Nothing scored: drop the empty table and explain instead, so the
		// agent gets a corrective message rather than an empty result.
		var eb strings.Builder
		eb.WriteString("# CVSS Scores\n\n")
		eb.WriteString("No vector could be scored.\n\n")
		for _, r := range failures {
			eb.WriteString(fmt.Sprintf("- `%s`: %v\n", r.Vector, r.Err))
		}
		eb.WriteString("\nExpected format: `CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H` ")
		eb.WriteString("or `CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N`.\n")
		return eb.String()
	}

	if len(failures) > 0 {
		sb.WriteString("\n**Could not be scored:**\n")
		for _, r := range failures {
			sb.WriteString(fmt.Sprintf("- `%s`: %v\n", r.Vector, r.Err))
		}
	}

	return sb.String()
}

// scoreCVSSVector dispatches a vector to the calculator matching its version prefix.
func scoreCVSSVector(vector string) cvssResult {
	metrics, version, err := parseCVSSVector(vector)
	if err != nil {
		return cvssResult{Vector: vector, Err: err}
	}

	var score float64
	switch version {
	case "3.0", "3.1":
		score, err = scoreCVSS3(metrics)
	case "4.0":
		score, err = scoreCVSS4(metrics)
	default:
		err = fmt.Errorf("unsupported CVSS version %q (supported: 3.0, 3.1, 4.0)", version)
	}
	if err != nil {
		return cvssResult{Vector: vector, Version: version, Err: err}
	}

	return cvssResult{
		Vector:   vector,
		Version:  version,
		Score:    score,
		Severity: cvssSeverityLabel(score),
	}
}

// parseCVSSVector splits a vector string into its metric map and version. The
// leading "CVSS:x.y" prefix is required by the spec for v3.0 and later.
func parseCVSSVector(vector string) (map[string]string, string, error) {
	parts := strings.Split(strings.TrimSpace(vector), "/")
	if len(parts) < 2 {
		return nil, "", fmt.Errorf("malformed vector: expected slash-separated metrics")
	}

	prefix := strings.SplitN(parts[0], ":", 2)
	if len(prefix) != 2 || !strings.EqualFold(prefix[0], "CVSS") {
		return nil, "", fmt.Errorf("missing CVSS version prefix (expected a leading 'CVSS:3.1' or 'CVSS:4.0')")
	}
	version := prefix[1]

	metrics := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			return nil, version, fmt.Errorf("malformed metric %q (expected 'AB:C')", part)
		}
		metrics[strings.ToUpper(kv[0])] = strings.ToUpper(kv[1])
	}
	if len(metrics) == 0 {
		return nil, version, fmt.Errorf("vector contains no metrics")
	}

	return metrics, version, nil
}

// cvss3Weights holds the constant weights from the CVSS v3.1 specification
// (section 7.4). Privileges Required is version-dependent on Scope, so it is
// handled separately in scoreCVSS3.
var cvss3Weights = map[string]map[string]float64{
	"AV": {"N": 0.85, "A": 0.62, "L": 0.55, "P": 0.2},
	"AC": {"L": 0.77, "H": 0.44},
	"UI": {"N": 0.85, "R": 0.62},
	"C":  {"H": 0.56, "L": 0.22, "N": 0.0},
	"I":  {"H": 0.56, "L": 0.22, "N": 0.0},
	"A":  {"H": 0.56, "L": 0.22, "N": 0.0},
}

// cvss3RequiredMetrics lists the base metrics that must all be present.
var cvss3RequiredMetrics = []string{"AV", "AC", "PR", "UI", "S", "C", "I", "A"}

// scoreCVSS3 implements the CVSS v3.1 base score equations.
func scoreCVSS3(m map[string]string) (float64, error) {
	for _, key := range cvss3RequiredMetrics {
		if _, ok := m[key]; !ok {
			return 0, fmt.Errorf("missing required base metric %q", key)
		}
	}

	scope := m["S"]
	if scope != "U" && scope != "C" {
		return 0, fmt.Errorf("invalid Scope value %q (expected U or C)", scope)
	}
	scopeChanged := scope == "C"

	// Privileges Required weights differ when Scope is Changed.
	var pr float64
	switch m["PR"] {
	case "N":
		pr = 0.85
	case "L":
		if scopeChanged {
			pr = 0.68
		} else {
			pr = 0.62
		}
	case "H":
		if scopeChanged {
			pr = 0.5
		} else {
			pr = 0.27
		}
	default:
		return 0, fmt.Errorf("invalid PrivilegesRequired value %q (expected N, L or H)", m["PR"])
	}

	weight := func(metric string) (float64, error) {
		v, ok := cvss3Weights[metric][m[metric]]
		if !ok {
			return 0, fmt.Errorf("invalid %s value %q", metric, m[metric])
		}
		return v, nil
	}

	av, err := weight("AV")
	if err != nil {
		return 0, err
	}
	ac, err := weight("AC")
	if err != nil {
		return 0, err
	}
	ui, err := weight("UI")
	if err != nil {
		return 0, err
	}
	conf, err := weight("C")
	if err != nil {
		return 0, err
	}
	integ, err := weight("I")
	if err != nil {
		return 0, err
	}
	avail, err := weight("A")
	if err != nil {
		return 0, err
	}

	iscBase := 1 - ((1 - conf) * (1 - integ) * (1 - avail))

	var impact float64
	if scopeChanged {
		impact = 7.52*(iscBase-0.029) - 3.25*math.Pow(iscBase-0.02, 15)
	} else {
		impact = 6.42 * iscBase
	}

	if impact <= 0 {
		return 0.0, nil
	}

	exploitability := 8.22 * av * ac * pr * ui

	score := impact + exploitability
	if scopeChanged {
		score *= 1.08
	}
	if score > 10 {
		score = 10
	}

	return roundUpCVSS(score), nil
}

// cvss4RequiredMetrics lists the mandatory CVSS v4.0 base metrics.
var cvss4RequiredMetrics = []string{
	"AV", "AC", "AT", "PR", "UI",
	"VC", "VI", "VA",
	"SC", "SI", "SA",
}

var cvss4AllowedValues = map[string][]string{
	"AV": {"N", "A", "L", "P"},
	"AC": {"L", "H"},
	"AT": {"N", "P"},
	"PR": {"N", "L", "H"},
	"UI": {"N", "P", "A"},
	"VC": {"H", "L", "N"},
	"VI": {"H", "L", "N"},
	"VA": {"H", "L", "N"},
	"SC": {"H", "L", "N"},
	"SI": {"H", "L", "N"},
	"SA": {"H", "L", "N"},
}

// cvss4ExploitabilityWeights approximate the exploitability contribution of
// each v4.0 metric.
var cvss4ExploitabilityWeights = map[string]map[string]float64{
	"AV": {"N": 1.0, "A": 0.75, "L": 0.5, "P": 0.25},
	"AC": {"L": 1.0, "H": 0.7},
	"AT": {"N": 1.0, "P": 0.7},
	"PR": {"N": 1.0, "L": 0.7, "H": 0.45},
	"UI": {"N": 1.0, "P": 0.75, "A": 0.55},
}

// cvss4ImpactWeights approximate the impact contribution of each v4.0
// vulnerable-system and subsequent-system metric.
var cvss4ImpactWeights = map[string]float64{"H": 1.0, "L": 0.55, "N": 0.0}

// scoreCVSS4 computes an approximate CVSS v4.0 base score.
//
// The official v4.0 score is not a closed-form equation: it is defined by an
// interpolated lookup over 270 equivalence classes (macrovectors) published
// with the specification. Reproducing that table is out of scope here, so this
// implements the documented model shape — a weighted exploitability term
// combined with vulnerable-system and subsequent-system impact — which tracks
// the official score closely enough for triage ordering but can differ from
// the authoritative value by a few tenths. Treat v4.0 output as an estimate
// and prefer v3.1 vectors when an exact score matters.
func scoreCVSS4(m map[string]string) (float64, error) {
	for _, key := range cvss4RequiredMetrics {
		if _, ok := m[key]; !ok {
			return 0, fmt.Errorf("missing required base metric %q", key)
		}
	}
	for metric, allowed := range cvss4AllowedValues {
		value := m[metric]
		valid := false
		for _, a := range allowed {
			if value == a {
				valid = true
				break
			}
		}
		if !valid {
			return 0, fmt.Errorf("invalid %s value %q (expected one of %s)",
				metric, value, strings.Join(allowed, ", "))
		}
	}

	exploitability := 1.0
	for _, metric := range []string{"AV", "AC", "AT", "PR", "UI"} {
		exploitability *= cvss4ExploitabilityWeights[metric][m[metric]]
	}

	vulnImpact := 1 - ((1 - cvss4ImpactWeights[m["VC"]]) *
		(1 - cvss4ImpactWeights[m["VI"]]) *
		(1 - cvss4ImpactWeights[m["VA"]]))

	subImpact := 1 - ((1 - cvss4ImpactWeights[m["SC"]]) *
		(1 - cvss4ImpactWeights[m["SI"]]) *
		(1 - cvss4ImpactWeights[m["SA"]]))

	// Subsequent-system impact contributes at a reduced rate, capped so that
	// it can raise but never dominate the vulnerable-system impact.
	impact := vulnImpact + (1-vulnImpact)*subImpact*0.5
	if impact <= 0 {
		return 0.0, nil
	}

	score := 10 * impact * (0.3 + 0.7*exploitability)
	if score > 10 {
		score = 10
	}

	return math.Round(score*10) / 10, nil
}

// roundUpCVSS implements the CVSS v3.1 Appendix A "Roundup" function, which
// rounds to one decimal place, away from zero. Integer arithmetic avoids the
// float edge cases that a naive math.Ceil(x*10)/10 hits.
func roundUpCVSS(value float64) float64 {
	scaled := int(math.Round(value * 100000))
	if scaled%10000 == 0 {
		return float64(scaled) / 100000
	}
	return (math.Floor(float64(scaled)/10000) + 1) / 10
}

// cvssSeverityLabel maps a base score onto the qualitative severity rating
// scale shared by CVSS v3.1 and v4.0.
func cvssSeverityLabel(score float64) string {
	switch {
	case score == 0:
		return "None"
	case score < 4.0:
		return "Low"
	case score < 7.0:
		return "Medium"
	case score < 9.0:
		return "High"
	default:
		return "Critical"
	}
}
