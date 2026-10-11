package mcpbridge

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestMCPToolRateLimiterSeparatesPrincipalsAndToolClasses(t *testing.T) {
	now := time.Unix(100, 0)
	limiter := newMCPToolRateLimiterWithClock(2, 1, func() time.Time { return now })
	principalA := context.WithValue(context.Background(), mcpPrincipalHashContextKey{}, "principal-a")
	principalB := context.WithValue(context.Background(), mcpPrincipalHashContextKey{}, "principal-b")

	for call := 0; call < 2; call++ {
		if allowed, _ := limiter.Allow(principalA, mcpToolGetFlowStatus); !allowed {
			t.Fatalf("read call %d for principal A was unexpectedly denied", call+1)
		}
	}
	if allowed, retryAfter := limiter.Allow(principalA, mcpToolGetFlowStatus); allowed || retryAfter <= 0 {
		t.Fatalf("third read call = allowed %t, retryAfter %s; want denial", allowed, retryAfter)
	}
	if allowed, _ := limiter.Allow(principalB, mcpToolGetFlowStatus); !allowed {
		t.Fatal("principal B should have an independent read budget")
	}
	if allowed, _ := limiter.Allow(principalA, mcpToolSubmitFlowInput); !allowed {
		t.Fatal("write tool should have an independent budget")
	}
	if allowed, _ := limiter.Allow(principalA, mcpToolSubmitFlowInput); allowed {
		t.Fatal("second write call should exceed the one-call budget")
	}

	now = now.Add(30 * time.Second)
	if allowed, _ := limiter.Allow(principalA, mcpToolGetFlowStatus); !allowed {
		t.Fatal("read bucket did not refill after thirty seconds")
	}
}

func TestMCPToolRateLimiterZeroDisablesClass(t *testing.T) {
	limiter := newMCPToolRateLimiterWithClock(0, 0, time.Now)
	for _, toolName := range []string{mcpToolGetFlowStatus, mcpToolStopFlow} {
		if allowed, retryAfter := limiter.Allow(context.Background(), toolName); !allowed || retryAfter != 0 {
			t.Fatalf("zero limit for %q = allowed %t, retryAfter %s; want unlimited", toolName, allowed, retryAfter)
		}
	}
}

func TestNormalizeMCPApprovalMode(t *testing.T) {
	testCases := map[string]string{
		"scope":         MCPApprovalModeScope,
		" DESTRUCTIVE ": MCPApprovalModeDestructive,
		"write":         MCPApprovalModeWrite,
		"unknown":       DefaultMCPApprovalMode,
		"":              DefaultMCPApprovalMode,
	}
	for input, expected := range testCases {
		if got := NormalizeMCPApprovalMode(input); got != expected {
			t.Errorf("NormalizeMCPApprovalMode(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestApprovalMCPToolCall(t *testing.T) {
	nextCalls := 0
	next := func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nextCalls++
		return mcp.NewToolResultText("executed"), nil
	}

	destructiveHandler := approvalMCPToolCall(MCPApprovalModeDestructive)(next)
	result, err := destructiveHandler(withMCPAuditState(context.Background()), mcp.CallToolRequest{Params: mcp.CallToolParams{Name: mcpToolGetFlowStatus}})
	if err != nil || result == nil || result.IsError || nextCalls != 1 {
		t.Fatalf("read call was not passed: result=%#v err=%v nextCalls=%d", result, err, nextCalls)
	}

	deniedContext := withMCPAuditState(context.Background())
	result, err = destructiveHandler(deniedContext, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: mcpToolStopFlow}})
	if err != nil || result == nil || !result.IsError || nextCalls != 1 {
		t.Fatalf("unapproved destructive call = result=%#v err=%v nextCalls=%d", result, err, nextCalls)
	}
	if got := mcpAuditDecisionFromContext(deniedContext); got != mcpAuditDecisionApproval {
		t.Fatalf("approval decision = %q, want %s", got, mcpAuditDecisionApproval)
	}

	approved := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: mcpToolStopFlow}, Header: make(http.Header)}
	approved.Header.Set(mcpApprovalHeader, mcpApprovalValue)
	result, err = destructiveHandler(context.Background(), approved)
	if err != nil || result == nil || result.IsError || nextCalls != 2 {
		t.Fatalf("approved destructive call was not passed: result=%#v err=%v nextCalls=%d", result, err, nextCalls)
	}
}

func TestRateLimitMCPToolCallRecordsDenial(t *testing.T) {
	limiter := newMCPToolRateLimiterWithClock(1, 1, func() time.Time { return time.Unix(100, 0) })
	next := func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("executed"), nil
	}
	handler := rateLimitMCPToolCall(limiter)(next)
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: mcpToolGetFlowStatus}}
	if _, err := handler(withMCPAuditState(context.Background()), request); err != nil {
		t.Fatalf("first rate-limit call returned error: %v", err)
	}
	deniedContext := withMCPAuditState(context.Background())
	if result, err := handler(deniedContext, request); err != nil || result == nil || !result.IsError {
		t.Fatalf("second rate-limit call = result=%#v err=%v, want tool error", result, err)
	}
	if got := mcpAuditDecisionFromContext(deniedContext); got != mcpAuditDecisionRateLimited {
		t.Fatalf("rate-limit decision = %q, want %s", got, mcpAuditDecisionRateLimited)
	}
}

func TestMCPRequestContextMiddlewareAddsCorrelationAndClientContext(t *testing.T) {
	var requestID string
	var clientIP string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID = mcpRequestIDFromContext(r.Context())
		clientIP = mcpClientIPFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/mcp/message", nil)
	request.RemoteAddr = "192.0.2.10:4567"
	MCPRequestContextMiddleware(next).ServeHTTP(recorder, request)

	if requestID == "" {
		t.Fatal("request ID was not added to context")
	}
	if recorder.Header().Get("X-Request-ID") != requestID {
		t.Fatalf("response request ID = %q, context request ID = %q", recorder.Header().Get("X-Request-ID"), requestID)
	}
	if clientIP != "192.0.2.10" {
		t.Fatalf("client IP = %q, want 192.0.2.10", clientIP)
	}
}
