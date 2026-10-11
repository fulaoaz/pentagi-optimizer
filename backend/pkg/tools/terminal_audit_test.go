package tools

import (
	"context"
	"testing"
	"time"

	obs "pentagi/pkg/observability"
)

func TestMCPSandboxActionAuditFieldsKeepOnlySafeMetadata(t *testing.T) {
	taskID := int64(22)
	subtaskID := int64(33)
	ctx := obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     "request-123",
		PrincipalHash: "principal-hash",
		SessionHash:   "session-hash",
		OriginHash:    "origin-hash",
		ClientIPHash:  "client-ip-hash",
	})

	fields, ok := mcpSandboxActionAuditFields(
		ctx,
		11,
		&taskID,
		&subtaskID,
		TerminalToolName,
		"command",
		"success",
		250*time.Millisecond,
	)
	if !ok {
		t.Fatal("MCP correlation did not produce sandbox audit fields")
	}

	for key, expected := range map[string]any{
		"action":         "mcp_sandbox_action",
		"tool":           TerminalToolName,
		"operation":      "command",
		"outcome":        "success",
		"duration_ms":    int64(250),
		"flow_id":        int64(11),
		"task_id":        taskID,
		"subtask_id":     subtaskID,
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

	for _, forbidden := range []string{"args", "input", "command", "cwd", "path", "result", "output"} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("sandbox audit fields must not include %q", forbidden)
		}
	}

	if _, ok := mcpSandboxActionAuditFields(context.Background(), 11, &taskID, &subtaskID, TerminalToolName, "command", "success", 0); ok {
		t.Fatal("non-MCP context unexpectedly produced sandbox audit fields")
	}
}
