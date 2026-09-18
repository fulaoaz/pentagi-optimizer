package mcpbridge

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	obs "pentagi/pkg/observability"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mcpRequestIDContextKey struct{}
type mcpClientIPContextKey struct{}
type mcpPrincipalHashContextKey struct{}
type mcpAuditStateContextKey struct{}

const (
	DefaultMCPReadToolRateLimit  = 60
	DefaultMCPWriteToolRateLimit = 10
	DefaultMCPApprovalMode       = "scope"
	MCPApprovalModeScope         = "scope"
	MCPApprovalModeDestructive   = "destructive"
	MCPApprovalModeWrite         = "write"
	mcpAuditDecisionAllowed      = "allowed"
	mcpAuditDecisionRateLimited  = "rate_limited"
	mcpAuditDecisionApproval     = "approval_required"
	mcpAuditDecisionWriteScope   = "write_scope_required"
	mcpAuditDecisionOutputLimit  = "output_limited"
	mcpAuditDecisionOutputEncode = "output_encoding_error"
	mcpApprovalHeader            = "X-MCP-Approval"
	mcpApprovalValue             = "confirm"
	maxMCPRateLimitBuckets       = 4096
)

// MCPGovernanceConfig controls MCP tool rate limits and approval gates.
type MCPGovernanceConfig struct {
	ReadRequestsPerMinute  int
	WriteRequestsPerMinute int
	ApprovalMode           string
}

// DefaultMCPGovernanceConfig returns the production-safe MCP governance defaults.
func DefaultMCPGovernanceConfig() MCPGovernanceConfig {
	return MCPGovernanceConfig{
		ReadRequestsPerMinute:  DefaultMCPReadToolRateLimit,
		WriteRequestsPerMinute: DefaultMCPWriteToolRateLimit,
		ApprovalMode:           DefaultMCPApprovalMode,
	}
}

// NormalizeMCPApprovalMode converts an unknown approval mode to the scope mode.
func NormalizeMCPApprovalMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case MCPApprovalModeDestructive:
		return MCPApprovalModeDestructive
	case MCPApprovalModeWrite:
		return MCPApprovalModeWrite
	case MCPApprovalModeScope:
		return MCPApprovalModeScope
	default:
		return DefaultMCPApprovalMode
	}
}

func normalizeMCPGovernanceConfig(config MCPGovernanceConfig) MCPGovernanceConfig {
	if config.ReadRequestsPerMinute < 0 {
		config.ReadRequestsPerMinute = DefaultMCPReadToolRateLimit
	}
	if config.WriteRequestsPerMinute < 0 {
		config.WriteRequestsPerMinute = DefaultMCPWriteToolRateLimit
	}
	config.ApprovalMode = NormalizeMCPApprovalMode(config.ApprovalMode)
	return config
}

type mcpRateLimitBucket struct {
	tokens    float64
	updatedAt time.Time
	lastSeen  time.Time
}

// MCPToolRateLimiter is a bounded, per-principal and per-tool token bucket.
type MCPToolRateLimiter struct {
	mu         sync.Mutex
	readLimit  int
	writeLimit int
	maxEntries int
	now        func() time.Time
	buckets    map[string]mcpRateLimitBucket
}

// NewMCPToolRateLimiter creates a limiter. A zero limit disables that class.
func NewMCPToolRateLimiter(readRequestsPerMinute, writeRequestsPerMinute int) *MCPToolRateLimiter {
	return &MCPToolRateLimiter{
		readLimit:  normalizeMCPRateLimitValue(readRequestsPerMinute, DefaultMCPReadToolRateLimit),
		writeLimit: normalizeMCPRateLimitValue(writeRequestsPerMinute, DefaultMCPWriteToolRateLimit),
		maxEntries: maxMCPRateLimitBuckets,
		now:        time.Now,
		buckets:    make(map[string]mcpRateLimitBucket),
	}
}

func newMCPToolRateLimiterWithClock(readRequestsPerMinute, writeRequestsPerMinute int, now func() time.Time) *MCPToolRateLimiter {
	limiter := NewMCPToolRateLimiter(readRequestsPerMinute, writeRequestsPerMinute)
	if now != nil {
		limiter.now = now
	}
	return limiter
}

func normalizeMCPRateLimitValue(value, defaultValue int) int {
	if value < 0 {
		return defaultValue
	}
	return value
}

// Allow consumes one token for a tool and returns a retry duration when denied.
func (l *MCPToolRateLimiter) Allow(ctx context.Context, toolName string) (bool, time.Duration) {
	if l == nil || strings.TrimSpace(toolName) == "" {
		return true, 0
	}

	limit := l.readLimit
	if mcpToolIsWrite(toolName) {
		limit = l.writeLimit
	}
	if limit <= 0 {
		return true, 0
	}

	now := l.now()
	key := mcpRateLimitIdentity(ctx) + "\x00" + toolName
	ratePerSecond := float64(limit) / 60.0
	capacity := float64(limit)

	l.mu.Lock()
	defer l.mu.Unlock()
	l.evictStale(now)

	bucket, exists := l.buckets[key]
	if !exists {
		bucket = mcpRateLimitBucket{tokens: capacity, updatedAt: now, lastSeen: now}
	}
	if now.Before(bucket.updatedAt) {
		bucket.updatedAt = now
	}
	if elapsed := now.Sub(bucket.updatedAt); elapsed > 0 {
		bucket.tokens = math.Min(capacity, bucket.tokens+elapsed.Seconds()*ratePerSecond)
		bucket.updatedAt = now
	}
	bucket.lastSeen = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		l.buckets[key] = bucket
		return true, 0
	}

	retryAfter := time.Duration(math.Ceil((1 - bucket.tokens) / ratePerSecond * float64(time.Second)))
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	l.buckets[key] = bucket
	return false, retryAfter
}

func (l *MCPToolRateLimiter) evictStale(now time.Time) {
	if len(l.buckets) < l.maxEntries {
		return
	}

	cutoff := now.Add(-2 * time.Minute)
	for key, bucket := range l.buckets {
		if bucket.lastSeen.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
	if len(l.buckets) < l.maxEntries {
		return
	}

	var oldestKey string
	var oldest time.Time
	for key, bucket := range l.buckets {
		if oldestKey == "" || bucket.lastSeen.Before(oldest) {
			oldestKey = key
			oldest = bucket.lastSeen
		}
	}
	if oldestKey != "" {
		delete(l.buckets, oldestKey)
	}
}

func rateLimitMCPToolCall(limiter *MCPToolRateLimiter) server.ToolHandlerMiddleware {
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if limiter == nil {
				return next(ctx, request)
			}
			allowed, retryAfter := limiter.Allow(ctx, request.Params.Name)
			if !allowed {
				setMCPAuditDecision(ctx, mcpAuditDecisionRateLimited)
				seconds := int(math.Ceil(retryAfter.Seconds()))
				return mcp.NewToolResultError(fmt.Sprintf("tool rate limit exceeded; retry after %d seconds", seconds)), nil
			}
			return next(ctx, request)
		}
	}
}

func approvalMCPToolCall(mode string) server.ToolHandlerMiddleware {
	mode = NormalizeMCPApprovalMode(mode)
	return func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			if !mcpApprovalRequired(mode, request.Params.Name) || mcpExplicitApprovalPresent(request) {
				return next(ctx, request)
			}
			setMCPAuditDecision(ctx, mcpAuditDecisionApproval)
			return mcp.NewToolResultError(fmt.Sprintf("explicit approval required; set %s: %s", mcpApprovalHeader, mcpApprovalValue)), nil
		}
	}
}

func mcpApprovalRequired(mode, toolName string) bool {
	switch NormalizeMCPApprovalMode(mode) {
	case MCPApprovalModeDestructive:
		return mcpToolRiskLevel(toolName) == "destructive"
	case MCPApprovalModeWrite:
		return mcpToolIsWrite(toolName)
	default:
		return false
	}
}

func mcpExplicitApprovalPresent(request mcp.CallToolRequest) bool {
	return strings.EqualFold(strings.TrimSpace(request.Header.Get(mcpApprovalHeader)), mcpApprovalValue)
}

func MCPRequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r == nil {
			next.ServeHTTP(w, r)
			return
		}

		requestID := uuid.NewString()
		ctx := context.WithValue(r.Context(), mcpRequestIDContextKey{}, requestID)
		if clientIP := mcpRemoteIP(r.RemoteAddr); clientIP != "" {
			ctx = context.WithValue(ctx, mcpClientIPContextKey{}, clientIP)
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ensureMCPRequestID(ctx context.Context) context.Context {
	if mcpRequestIDFromContext(ctx) != "" {
		return ctx
	}
	return context.WithValue(ctx, mcpRequestIDContextKey{}, uuid.NewString())
}

func mcpRemoteIP(remoteAddr string) string {
	remoteAddr = strings.TrimSpace(remoteAddr)
	if remoteAddr == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}

func withMCPPrincipalHash(r *http.Request, token string) *http.Request {
	if r == nil || strings.TrimSpace(token) == "" {
		return r
	}
	ctx := context.WithValue(r.Context(), mcpPrincipalHashContextKey{}, mcpIdentifierHash(token))
	return r.WithContext(ctx)
}

func mcpRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(mcpRequestIDContextKey{}).(string)
	return strings.TrimSpace(requestID)
}

func mcpPrincipalHashFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	principalHash, _ := ctx.Value(mcpPrincipalHashContextKey{}).(string)
	return strings.TrimSpace(principalHash)
}

func mcpClientIPFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	clientIP, _ := ctx.Value(mcpClientIPContextKey{}).(string)
	return strings.TrimSpace(clientIP)
}

func mcpRateLimitIdentity(ctx context.Context) string {
	if principalHash := mcpPrincipalHashFromContext(ctx); principalHash != "" {
		return "principal:" + principalHash
	}
	if clientIP := mcpClientIPFromContext(ctx); clientIP != "" {
		return "ip:" + mcpIdentifierHash(clientIP)
	}
	if ctx != nil {
		if session := server.ClientSessionFromContext(ctx); session != nil && session.SessionID() != "" {
			return "session:" + mcpIdentifierHash(session.SessionID())
		}
	}
	return "anonymous"
}

func mcpIdentifierHash(value string) string {
	return obs.CorrelationHash(value)
}

func mcpAuditCorrelation(ctx context.Context) obs.AuditCorrelation {
	return obs.AuditCorrelation{
		Source:        obs.AuditSourceMCP,
		RequestID:     mcpRequestIDFromContext(ctx),
		PrincipalHash: mcpPrincipalHashFromContext(ctx),
		SessionHash:   mcpSessionIDHash(ctx),
		OriginHash:    mcpOriginHashFromContext(ctx),
		ClientIPHash:  mcpClientIPHashFromContext(ctx),
	}
}

func mcpClientIPHashFromContext(ctx context.Context) string {
	return mcpIdentifierHash(mcpClientIPFromContext(ctx))
}

func mcpSessionIDHash(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if session := server.ClientSessionFromContext(ctx); session != nil && session.SessionID() != "" {
		return mcpIdentifierHash(session.SessionID())
	}
	return ""
}

type mcpAuditState struct {
	decision string
}

func withMCPAuditState(ctx context.Context) context.Context {
	return context.WithValue(ctx, mcpAuditStateContextKey{}, &mcpAuditState{decision: mcpAuditDecisionAllowed})
}

func setMCPAuditDecision(ctx context.Context, decision string) {
	if ctx == nil {
		return
	}
	state, _ := ctx.Value(mcpAuditStateContextKey{}).(*mcpAuditState)
	if state != nil && strings.TrimSpace(decision) != "" {
		state.decision = decision
	}
}

func mcpAuditDecisionFromContext(ctx context.Context) string {
	if ctx == nil {
		return mcpAuditDecisionAllowed
	}
	state, _ := ctx.Value(mcpAuditStateContextKey{}).(*mcpAuditState)
	if state == nil || strings.TrimSpace(state.decision) == "" {
		return mcpAuditDecisionAllowed
	}
	return state.decision
}
