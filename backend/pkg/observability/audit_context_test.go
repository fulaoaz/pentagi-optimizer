package observability

import (
	"context"
	"testing"
)

func TestAuditCorrelationContextKeepsOnlySanitizedIdentifiers(t *testing.T) {
	want := AuditCorrelation{
		Source:        AuditSourceMCP,
		RequestID:     "request-123",
		PrincipalHash: "principal-hash",
		SessionHash:   "session-hash",
		OriginHash:    "origin-hash",
		ClientIPHash:  "client-ip-hash",
	}
	got := AuditCorrelationFromContext(WithAuditCorrelation(context.Background(), want))
	if got != want {
		t.Fatalf("correlation = %#v, want %#v", got, want)
	}

	fields := got.Fields()
	for key, expected := range map[string]any{
		"audit_source":   AuditSourceMCP,
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
	for _, forbidden := range []string{"input", "arguments", "result", "authorization"} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("correlation fields must not include %q", forbidden)
		}
	}
}

func TestCorrelationHashIsStableAndDoesNotExposeValue(t *testing.T) {
	const value = "sensitive flow input"
	first := CorrelationHash(value)
	if first == "" || first == value {
		t.Fatalf("correlation hash = %q, want a non-empty fingerprint", first)
	}
	if second := CorrelationHash(value); second != first {
		t.Fatalf("correlation hashes differ: %q and %q", first, second)
	}
	if other := CorrelationHash("different input"); other == first {
		t.Fatalf("different values share correlation hash %q", first)
	}
	if got := CorrelationHash(""); got != "" {
		t.Fatalf("empty value hash = %q, want empty", got)
	}
}
