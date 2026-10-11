package docker

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"pentagi/pkg/database"
	obs "pentagi/pkg/observability"
)

func TestContainerExitAuditFieldsKeepOnlySafeMetadata(t *testing.T) {
	ctx := WithContainerAuditContext(
		obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
			Source:        obs.AuditSourceMCP,
			RequestID:     "request-123",
			PrincipalHash: "principal-hash",
			SessionHash:   "session-hash",
			OriginHash:    "origin-hash",
			ClientIPHash:  "client-ip-hash",
		}),
		42,
		database.ContainerTypePrimary,
	)

	fields := containerExitAuditFields(ctx, "stop", "requested", "success", 275*time.Millisecond)
	for key, expected := range map[string]any{
		"action":         "container_exit",
		"operation":      "stop",
		"reason":         "requested",
		"outcome":        "success",
		"duration_ms":    int64(275),
		"flow_id":        int64(42),
		"container_type": database.ContainerTypePrimary,
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

	for _, forbidden := range []string{
		"container_id", "local_id", "container_name", "name", "image", "work_dir",
		"host_dir", "path", "error", "error_message", "output", "command",
	} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("container exit audit fields must not include %q", forbidden)
		}
	}
}

func TestContainerAuditTokensAreBounded(t *testing.T) {
	fields := containerExitAuditFields(
		WithContainerAuditContext(context.Background(), 9, database.ContainerType("secret-type")),
		"secret-operation",
		"secret-reason",
		"secret-outcome",
		-1,
	)

	if fields["operation"] != "other" || fields["reason"] != "other" || fields["outcome"] != "error" {
		t.Fatalf("untrusted audit tokens were not normalized: %#v", fields)
	}
	if fields["duration_ms"] != int64(0) {
		t.Fatalf("negative duration was not clamped: %v", fields["duration_ms"])
	}
	if _, found := fields["container_type"]; found {
		t.Fatal("invalid container type must not enter the audit record")
	}
}

func TestContainerCleanupAuditFieldsExposeCountsWithoutIdentifiers(t *testing.T) {
	ctx := obs.WithAuditCorrelation(context.Background(), obs.AuditCorrelation{
		Source:      obs.AuditSourceMCP,
		RequestID:   "request-cleanup",
		SessionHash: "session-cleanup",
	})
	fields := containerCleanupAuditFields(ctx, "error", 3, 7, 5, 2, 4*time.Second)

	for key, expected := range map[string]any{
		"action":          "container_cleanup",
		"outcome":         "error",
		"flow_count":      3,
		"attempted_count": 7,
		"succeeded_count": 5,
		"failed_count":    2,
		"duration_ms":     int64(4000),
		"audit_source":    obs.AuditSourceMCP,
		"request_id":      "request-cleanup",
		"session_hash":    "session-cleanup",
	} {
		if actual := fields[key]; actual != expected {
			t.Fatalf("field %q = %v, want %v", key, actual, expected)
		}
	}

	serialized := fmt.Sprint(fields)
	for _, sensitive := range []string{"container-id", "container-name", "docker://daemon", "secret-error"} {
		if strings.Contains(serialized, sensitive) {
			t.Fatalf("cleanup audit fields contain sensitive value %q: %s", sensitive, serialized)
		}
	}
}
