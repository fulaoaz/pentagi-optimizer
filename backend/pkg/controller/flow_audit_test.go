package controller

import (
	"context"
	"testing"

	obs "pentagi/pkg/observability"
)

func TestNewFlowInputCopiesAuditCorrelationWithoutRequestContext(t *testing.T) {
	correlation := obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-123",
		PrincipalHash: "principal-hash",
		SessionHash:   "session-hash",
		OriginHash:    "origin-hash",
		ClientIPHash:  "client-ip-hash",
	}
	requestCtx, cancel := context.WithCancel(obs.WithAuditCorrelation(context.Background(), correlation))
	input := newFlowInput(requestCtx, "sensitive flow input")
	cancel()

	if input.correlation != correlation {
		t.Fatalf("input correlation = %#v, want %#v", input.correlation, correlation)
	}
	if input.done == nil {
		t.Fatal("input completion channel was not initialized")
	}
}

func TestMCPTaskTriggerAuditFieldsExcludeInputContent(t *testing.T) {
	ctx := obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-123",
		PrincipalHash: "principal-hash",
		SessionHash:   "session-hash",
		OriginHash:    "origin-hash",
		ClientIPHash:  "client-ip-hash",
	})
	fields, ok := mcpTaskTriggerAuditFields(ctx, 42, 84, "created")
	if !ok {
		t.Fatal("MCP correlation did not produce task audit fields")
	}
	for key, expected := range map[string]any{
		"action":         "mcp_task_trigger",
		"trigger":        "created",
		"flow_id":        int64(42),
		"task_id":        int64(84),
		"audit_source":   obs.AuditSourceMCP,
		"request_id":     "request-123",
		"principal_hash": "principal-hash",
		"session_hash":   "session-hash",
		"origin_hash":    "origin-hash",
		"client_ip_hash": "client-ip-hash",
	} {
		if actual := fields[key]; actual != expected {
			t.Fatalf("field %q = %v, want %v", key, actual, expected)
		}
	}
	for _, forbidden := range []string{"input", "input_hash", "task_title", "task_result"} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("task audit fields must not include %q", forbidden)
		}
	}
	if _, ok := mcpTaskTriggerAuditFields(context.Background(), 42, 84, "created"); ok {
		t.Fatal("non-MCP context unexpectedly produced task audit fields")
	}
}

func TestSafeFlowInputLogFieldsRedactInputContent(t *testing.T) {
	const input = "sensitive flow input"
	fields := safeFlowInputLogFields(input)
	if _, found := fields["input"]; found {
		t.Fatal("flow worker log fields must not contain raw input")
	}
	if got := fields["input_bytes"]; got != len(input) {
		t.Fatalf("input_bytes = %v, want %d", got, len(input))
	}
	if _, found := fields["input_hash"]; found {
		t.Fatal("flow worker log fields must not contain an input fingerprint")
	}
}
