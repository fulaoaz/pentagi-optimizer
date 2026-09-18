package observability

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type auditCorrelationContextKey struct{}

const AuditSourceMCP = "mcp"

// AuditCorrelation carries only stable, non-secret identifiers between an
// incoming request and asynchronous work.
type AuditCorrelation struct {
	Source        string
	RequestID     string
	PrincipalHash string
	SessionHash   string
	OriginHash    string
	ClientIPHash  string
}

// WithAuditCorrelation copies sanitized audit identifiers into a context.
func WithAuditCorrelation(ctx context.Context, correlation AuditCorrelation) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	correlation = correlation.normalized()
	if correlation.isZero() {
		return ctx
	}
	return context.WithValue(ctx, auditCorrelationContextKey{}, correlation)
}

// AuditCorrelationFromContext returns the sanitized correlation identifiers.
func AuditCorrelationFromContext(ctx context.Context) AuditCorrelation {
	if ctx == nil {
		return AuditCorrelation{}
	}
	correlation, _ := ctx.Value(auditCorrelationContextKey{}).(AuditCorrelation)
	return correlation.normalized()
}

// Fields returns the correlation identifiers safe to include in structured logs.
func (c AuditCorrelation) Fields() map[string]any {
	fields := make(map[string]any, 6)
	if c.Source != "" {
		fields["audit_source"] = c.Source
	}
	if c.RequestID != "" {
		fields["request_id"] = c.RequestID
	}
	if c.PrincipalHash != "" {
		fields["principal_hash"] = c.PrincipalHash
	}
	if c.SessionHash != "" {
		fields["session_hash"] = c.SessionHash
	}
	if c.OriginHash != "" {
		fields["origin_hash"] = c.OriginHash
	}
	if c.ClientIPHash != "" {
		fields["client_ip_hash"] = c.ClientIPHash
	}
	return fields
}

// CorrelationHash returns a short SHA-256 fingerprint for log correlation.
func CorrelationHash(value string) string {
	if value == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:8])
}

func (c AuditCorrelation) normalized() AuditCorrelation {
	c.Source = strings.ToLower(strings.TrimSpace(c.Source))
	c.RequestID = strings.TrimSpace(c.RequestID)
	c.PrincipalHash = strings.TrimSpace(c.PrincipalHash)
	c.SessionHash = strings.TrimSpace(c.SessionHash)
	c.OriginHash = strings.TrimSpace(c.OriginHash)
	c.ClientIPHash = strings.TrimSpace(c.ClientIPHash)
	return c
}

func (c AuditCorrelation) isZero() bool {
	return c.Source == "" && c.RequestID == "" && c.PrincipalHash == "" && c.SessionHash == "" && c.OriginHash == "" && c.ClientIPHash == ""
}
