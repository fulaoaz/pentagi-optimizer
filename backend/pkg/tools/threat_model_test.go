package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func threatModelArgs(t *testing.T, action string, fields map[string]string) json.RawMessage {
	t.Helper()
	obj := map[string]any{"action": action}
	for k, v := range fields {
		obj[k] = v
	}
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}
	return data
}

func TestThreatModelGetBeforeDeriveReturnsGuidance(t *testing.T) {
	tool := NewThreatModelTool(990001, nil)
	res, err := tool.Handle(context.Background(), ThreatModelToolName, threatModelArgs(t, "get", nil))
	if err != nil {
		t.Fatalf("get before derive should not error: %v", err)
	}
	if !strings.Contains(res, "No threat model") {
		t.Fatalf("expected guidance text, got: %s", res)
	}
}

func TestThreatModelDeriveRequiresAllSections(t *testing.T) {
	tool := NewThreatModelTool(990002, nil)
	_, err := tool.Handle(context.Background(), ThreatModelToolName, threatModelArgs(t, "derive", map[string]string{
		"overview": "only overview",
	}))
	if err == nil {
		t.Fatal("derive with a single section must fail")
	}
}

func TestThreatModelAmendBeforeDeriveFails(t *testing.T) {
	tool := NewThreatModelTool(990003, nil)
	_, err := tool.Handle(context.Background(), ThreatModelToolName, threatModelArgs(t, "amend", map[string]string{
		"amendment": "nudge scope",
	}))
	if err == nil {
		t.Fatal("amend before derive must fail")
	}
}

func TestThreatModelDeriveAmendGetFlow(t *testing.T) {
	flowID := int64(990004)
	tool := NewThreatModelTool(flowID, nil)
	ctx := context.Background()

	if _, err := tool.Handle(ctx, ThreatModelToolName, threatModelArgs(t, "derive", map[string]string{
		"overview":          "Single-host web target owned by the engagement sponsor.",
		"trust_boundaries":  "Public edge to internal API; no VPN hop.",
		"attack_surface":    "Three exposed HTTP routes and one admin panel.",
		"severity_criteria": "Critical equals unauthenticated data exposure.",
	})); err != nil {
		t.Fatalf("derive: %v", err)
	}

	if _, err := tool.Handle(ctx, ThreatModelToolName, threatModelArgs(t, "amend", map[string]string{
		"amendment": "Admin panel is behind an IP allowlist, treat it as second priority.",
	})); err != nil {
		t.Fatalf("amend: %v", err)
	}

	res, err := tool.Handle(ctx, ThreatModelToolName, threatModelArgs(t, "get", nil))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	for _, want := range []string{
		"Single-host web target",
		"trust_boundaries",
		"IP allowlist",
		"[UNTRUSTED FLOW DATA]",
	} {
		if !strings.Contains(res, want) {
			t.Fatalf("model response missing %q", want)
		}
	}

	other := NewThreatModelTool(flowID+1, nil)
	res2, err := other.Handle(ctx, ThreatModelToolName, threatModelArgs(t, "get", nil))
	if err != nil {
		t.Fatalf("other flow get: %v", err)
	}
	if strings.Contains(res2, "Single-host web target") {
		t.Fatal("threat model leaked across flows")
	}
}

func TestCoverageRecordAndListCounts(t *testing.T) {
	flowID := int64(990010)
	tool := NewCoverageTool(flowID, nil)
	ctx := context.Background()

	if res, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "list", nil)); err != nil {
		t.Fatalf("list before any record: %v", err)
	} else if !strings.Contains(res, "No coverage entries") {
		t.Fatalf("expected empty ledger text, got: %s", res)
	}

	records := []map[string]string{
		{"surface": "GET /api/v1/users (IDOR)", "outcome": "reported", "summary": "Cross-tenant read confirmed.", "title": "IDOR in user list endpoint", "location": "GET /api/v1/users?id=2 (user_id param)", "cvss_vector": "CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N"},
		{"surface": "GET /api/v1/orders", "outcome": "no_issue_found"},
		{"surface": "GraphQL introspection", "outcome": "ruled_out", "evidence": "Introspection returns schema=false on prod."},
		{"surface": "SAML SSO", "outcome": "not_applicable", "evidence": "Deployment uses local auth only."},
		{"surface": "Cache poisoning", "outcome": "needs_follow_up", "evidence": "Requires an origin probe the lab cannot reach."},
	}
	for _, r := range records {
		if _, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "record", r)); err != nil {
			t.Fatalf("record %s: %v", r["surface"], err)
		}
	}

	res, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "list", nil))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, want := range []string{
		"Coverage entries: 5",
		"reported: 1",
		"no_issue_found: 1",
		"ruled_out: 1",
		"not_applicable: 1",
		"needs_follow_up: 1",
		"cov-0001",
		"cov-0005",
	} {
		if !strings.Contains(res, want) {
			t.Fatalf("ledger response missing %q", want)
		}
	}

	if _, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "record", map[string]string{
		"surface": "z", "outcome": "reported",
	})); err == nil {
		t.Fatal("reported without title must fail")
	}

	res3, err3 := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "list", nil))
	if err3 != nil {
		t.Fatalf("list with findings: %v", err3)
	}
	for _, want := range []string{
		"finding: IDOR in user list endpoint",
		"location: GET /api/v1/users?id=2",
		"cvss: CVSS:3.1/AV:N/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N",
	} {
		if !strings.Contains(res3, want) {
			t.Fatalf("findings output missing %q", want)
		}
	}

	other := NewCoverageTool(flowID+1, nil)
	res2, _ := other.Handle(ctx, CoverageToolName, threatModelArgs(t, "list", nil))
	if !strings.Contains(res2, "No coverage entries") {
		t.Fatal("coverage ledger leaked across flows")
	}
}

func TestCoverageRecordValidation(t *testing.T) {
	tool := NewCoverageTool(990011, nil)
	ctx := context.Background()

	if _, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "record", map[string]string{
		"outcome": "reported",
	})); err == nil {
		t.Fatal("record without surface must fail")
	}

	if _, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "record", map[string]string{
		"surface": "x",
		"outcome": "maybe",
	})); err == nil {
		t.Fatal("record with unknown outcome must fail")
	}

	if _, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "record", map[string]string{
		"surface": "x",
		"outcome": "ruled_out",
	})); err == nil {
		t.Fatal("ruled_out without evidence must fail")
	}

	if _, err := tool.Handle(ctx, CoverageToolName, threatModelArgs(t, "record", map[string]string{
		"surface": "y",
		"outcome": "reported",
		"title":   "Open redirect on login",
	})); err != nil {
		t.Fatalf("reported with title but no evidence must succeed: %v", err)
	}
}

func TestThreatModelUnsupportedActionFails(t *testing.T) {
	tool := NewThreatModelTool(990012, nil)
	if _, err := tool.Handle(context.Background(), ThreatModelToolName, threatModelArgs(t, "delete", nil)); err == nil {
		t.Fatal("unsupported action must fail")
	}
}
