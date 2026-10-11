package mcpbridge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	obs "pentagi/pkg/observability"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

func TestFlowBridgeWriteToolsAreOptIn(t *testing.T) {
	testCases := []struct {
		name           string
		bridge         *FlowBridge
		wantWriteTools bool
	}{
		{
			name:           "default bridge is read only",
			bridge:         NewFlowBridge("PentAGI", "1.0.0", nil, nil),
			wantWriteTools: false,
		},
		{
			name:           "write tools can be enabled explicitly",
			bridge:         NewFlowBridgeWithOptions("PentAGI", "1.0.0", nil, nil, true),
			wantWriteTools: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			toolNames := listToolNames(t, testCase.bridge)

			for _, readTool := range []string{"get_flow_status", "list_assistants"} {
				if !toolNames[readTool] {
					t.Fatalf("missing read tool %q", readTool)
				}
			}
			for _, writeTool := range []string{"submit_flow_input", "stop_flow"} {
				if got := toolNames[writeTool]; got != testCase.wantWriteTools {
					t.Fatalf("tool %q enabled = %t, want %t", writeTool, got, testCase.wantWriteTools)
				}
			}
		})
	}
}

func TestFlowBridgeToolAllowlist(t *testing.T) {
	testCases := []struct {
		name          string
		allowedTools  []string
		wantToolNames map[string]bool
	}{
		{
			name: "empty allowlist preserves all enabled tools",
			wantToolNames: map[string]bool{
				mcpToolGetFlowStatus:   true,
				mcpToolSubmitFlowInput: true,
				mcpToolStopFlow:        true,
				mcpToolListAssistants:  true,
			},
		},
		{
			name:         "allowlist exposes only exact names",
			allowedTools: []string{" get_flow_status ", "stop_flow"},
			wantToolNames: map[string]bool{
				mcpToolGetFlowStatus: true,
				mcpToolStopFlow:      true,
			},
		},
		{
			name:         "allowlist can keep read tools without writes",
			allowedTools: []string{"get_flow_status,list_assistants"},
			wantToolNames: map[string]bool{
				mcpToolGetFlowStatus:  true,
				mcpToolListAssistants: true,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bridge := NewFlowBridgeWithToolPolicy("PentAGI", "1.0.0", nil, nil, true, nil, testCase.allowedTools)
			got := listToolNames(t, bridge)
			for name, want := range map[string]bool{
				mcpToolGetFlowStatus:   testCase.wantToolNames[mcpToolGetFlowStatus],
				mcpToolSubmitFlowInput: testCase.wantToolNames[mcpToolSubmitFlowInput],
				mcpToolStopFlow:        testCase.wantToolNames[mcpToolStopFlow],
				mcpToolListAssistants:  testCase.wantToolNames[mcpToolListAssistants],
			} {
				if got[name] != want {
					t.Fatalf("tool %q enabled = %t, want %t; tools = %v", name, got[name], want, got)
				}
			}
		})
	}
}

func TestNormalizedMCPToolAllowlist(t *testing.T) {
	testCases := []struct {
		name           string
		names          []string
		wantRestricted bool
		wantNames      []string
	}{
		{name: "empty", wantNames: nil},
		{name: "blank values", names: []string{" ", "\t"}, wantNames: nil},
		{name: "comma separated and deduplicated", names: []string{"get_flow_status, list_assistants", "get_flow_status"}, wantRestricted: true, wantNames: []string{"get_flow_status", "list_assistants"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got, restricted := normalizedMCPToolAllowlist(testCase.names)
			if restricted != testCase.wantRestricted {
				t.Fatalf("restricted = %t, want %t", restricted, testCase.wantRestricted)
			}
			if len(got) != len(testCase.wantNames) {
				t.Fatalf("allowlist length = %d, want %d (%v)", len(got), len(testCase.wantNames), got)
			}
			for _, name := range testCase.wantNames {
				if _, ok := got[name]; !ok {
					t.Fatalf("allowlist missing %q: %v", name, got)
				}
			}
		})
	}
}

func TestFlowBridgePublishesStrictToolSchemasAndSafetyHints(t *testing.T) {
	bridge := NewFlowBridgeWithOptions("PentAGI", "1.0.0", nil, nil, true)
	tools := bridge.GetServer().ListTools()

	expectedHints := map[string]struct {
		readOnly    bool
		destructive bool
	}{
		"get_flow_status":   {readOnly: true, destructive: false},
		"list_assistants":   {readOnly: true, destructive: false},
		"submit_flow_input": {readOnly: false, destructive: false},
		"stop_flow":         {readOnly: false, destructive: true},
	}

	for name, hints := range expectedHints {
		tool, ok := tools[name]
		if !ok {
			t.Fatalf("tool %q is missing", name)
		}
		if tool.Tool.InputSchema.AdditionalProperties != false {
			t.Fatalf("tool %q must reject unknown arguments", name)
		}
		if tool.Tool.Annotations.ReadOnlyHint == nil || *tool.Tool.Annotations.ReadOnlyHint != hints.readOnly {
			t.Fatalf("tool %q readOnlyHint is incorrect", name)
		}
		if tool.Tool.Annotations.DestructiveHint == nil || *tool.Tool.Annotations.DestructiveHint != hints.destructive {
			t.Fatalf("tool %q destructiveHint is incorrect", name)
		}
	}

	for _, name := range []string{"get_flow_status", "list_assistants", "submit_flow_input", "stop_flow"} {
		tool := tools[name]
		property, ok := tool.Tool.InputSchema.Properties["flow_id"].(map[string]any)
		if !ok {
			t.Fatalf("tool %q flow_id schema has type %T, want object", name, tool.Tool.InputSchema.Properties["flow_id"])
		}
		if property["type"] != "integer" {
			t.Fatalf("tool %q flow_id type = %v, want integer", name, property["type"])
		}
		if fmt.Sprint(property["minimum"]) != "1" {
			t.Fatalf("tool %q flow_id minimum = %v, want 1", name, property["minimum"])
		}
	}

	inputProperty, ok := tools["submit_flow_input"].Tool.InputSchema.Properties["input"].(map[string]any)
	if !ok {
		t.Fatalf("submit_flow_input input schema has type %T, want object", tools["submit_flow_input"].Tool.InputSchema.Properties["input"])
	}
	if inputProperty["minLength"] != 1 || fmt.Sprint(inputProperty["maxLength"]) != fmt.Sprint(maxMCPFlowInputLength) {
		t.Fatalf("input length constraints = min %v max %v", inputProperty["minLength"], inputProperty["maxLength"])
	}
}

func TestFlowIDBindingPreservesLargeIntegers(t *testing.T) {
	const expected int64 = 9007199254740993
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{
		RawArguments: json.RawMessage(`{"flow_id":9007199254740993}`),
	}}

	var args flowIDArguments
	if err := request.BindArguments(&args); err != nil {
		t.Fatalf("bind arguments: %v", err)
	}
	if args.FlowID != expected {
		t.Fatalf("flow_id = %d, want %d", args.FlowID, expected)
	}
}

func TestFlowBridgeRejectsInvalidArgumentsBeforeControllerAccess(t *testing.T) {
	bridge := NewFlowBridgeWithOptions("PentAGI", "1.0.0", nil, nil, true)

	testCases := []struct {
		name    string
		tool    string
		payload string
	}{
		{
			name:    "fractional flow id",
			tool:    "list_assistants",
			payload: `{"flow_id":1.5}`,
		},
		{
			name:    "unknown argument",
			tool:    "list_assistants",
			payload: `{"flow_id":1,"extra":"ignored"}`,
		},
		{
			name:    "missing conditional flow id",
			tool:    "get_flow_status",
			payload: `{"detail":"tasks"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			message := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":%q,"arguments":%s}}`, testCase.tool, testCase.payload)
			response := bridge.GetServer().HandleMessage(context.Background(), json.RawMessage(message))
			encoded, err := json.Marshal(response)
			if err != nil {
				t.Fatalf("marshal response: %v", err)
			}
			if !strings.Contains(string(encoded), `"isError":true`) {
				t.Fatalf("response = %s, want tool error", encoded)
			}
		})
	}
}

func TestMCPAuditDetailsExcludeArguments(t *testing.T) {
	ctx := context.WithValue(context.Background(), mcpAuthModeContextKey{}, mcpAuthModeBearer)
	ctx = context.WithValue(ctx, mcpClientIPContextKey{}, "192.0.2.10")
	ctx = context.WithValue(ctx, mcpOriginHashContextKey{}, mcpIdentifierHash("https://console.example.com"))
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "stop_flow",
			Arguments: map[string]any{
				"flow_id": 42,
				"input":   "sensitive flow input",
			},
		},
	}
	details := mcpAuditDetails(ctx, request, &mcp.CallToolResult{IsError: true}, errors.New("handler failed"), 125*time.Millisecond)

	if got := details["tool"]; got != "stop_flow" {
		t.Fatalf("tool = %v, want stop_flow", got)
	}
	if got := details["auth_mode"]; got != mcpAuthModeBearer {
		t.Fatalf("auth_mode = %v, want %s", got, mcpAuthModeBearer)
	}
	if got := details["access"]; got != "write" {
		t.Fatalf("access = %v, want write", got)
	}
	if got := details["outcome"]; got != "error" {
		t.Fatalf("outcome = %v, want error", got)
	}
	if got := details["duration_ms"]; got != int64(125) {
		t.Fatalf("duration_ms = %v, want 125", got)
	}
	if got := details["client_ip_hash"]; got != mcpIdentifierHash("192.0.2.10") {
		t.Fatalf("client_ip_hash = %v, want a hash of the client IP", got)
	}
	if got := details["origin_hash"]; got != mcpIdentifierHash("https://console.example.com") {
		t.Fatalf("origin_hash = %v, want a hash of the normalized origin", got)
	}
	if _, found := details["arguments"]; found {
		t.Fatal("audit details must not include tool arguments")
	}
	if _, found := details["input"]; found {
		t.Fatal("audit details must not include input fields")
	}
}

func TestAuditToolCallPropagatesSanitizedCorrelation(t *testing.T) {
	ctx := context.WithValue(context.Background(), mcpRequestIDContextKey{}, "request-123")
	ctx = context.WithValue(ctx, mcpPrincipalHashContextKey{}, "principal-hash")
	ctx = context.WithValue(ctx, mcpClientIPContextKey{}, "203.0.113.10")
	ctx = context.WithValue(ctx, mcpOriginHashContextKey{}, "origin-hash")

	var correlation obs.AuditCorrelation
	handler := auditToolCall(func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		correlation = obs.AuditCorrelationFromContext(ctx)
		return mcp.NewToolResultText("ok"), nil
	})
	if _, err := handler(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: mcpToolSubmitFlowInput}}); err != nil {
		t.Fatalf("audit handler returned error: %v", err)
	}

	if correlation.Source != obs.AuditSourceMCP {
		t.Fatalf("audit source = %q, want %q", correlation.Source, obs.AuditSourceMCP)
	}
	if correlation.RequestID != "request-123" || correlation.PrincipalHash != "principal-hash" || correlation.OriginHash != "origin-hash" {
		t.Fatalf("unexpected correlation: %#v", correlation)
	}
	if want := mcpIdentifierHash("203.0.113.10"); correlation.ClientIPHash != want {
		t.Fatalf("client IP hash = %q, want %q", correlation.ClientIPHash, want)
	}
	if correlation.ClientIPHash == "203.0.113.10" {
		t.Fatalf("client IP hash exposes the raw address: %#v", correlation)
	}
}

func TestAuthorizeMCPToolCallRequiresWriteScope(t *testing.T) {
	request := mcp.CallToolRequest{Params: mcp.CallToolParams{Name: "stop_flow"}}
	nextCalled := false
	next := func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nextCalled = true
		return mcp.NewToolResultText("executed"), nil
	}
	handler := authorizeMCPToolCall(next)

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unauthenticated authorization returned error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatal("unauthenticated write call was not rejected")
	}
	if nextCalled {
		t.Fatal("unauthenticated write call reached the handler")
	}

	readContext := withMCPAuditState(context.WithValue(context.Background(), mcpAccessScopeContextKey{}, mcpAccessScopeRead))
	result, err = handler(readContext, request)
	if err != nil {
		t.Fatalf("read-scoped authorization returned error: %v", err)
	}
	if result == nil || !result.IsError || nextCalled {
		t.Fatal("read-scoped write call was not rejected")
	}
	if got := mcpAuditDecisionFromContext(readContext); got != mcpAuditDecisionWriteScope {
		t.Fatalf("read-scoped audit decision = %q, want %s", got, mcpAuditDecisionWriteScope)
	}

	writeContext := context.WithValue(context.Background(), mcpAccessScopeContextKey{}, mcpAccessScopeWrite)
	result, err = handler(writeContext, request)
	if err != nil {
		t.Fatalf("write authorization returned error: %v", err)
	}
	if result == nil || result.IsError || !nextCalled {
		t.Fatal("write-scoped call did not reach the handler")
	}
}

func TestMCPOutputBuilderBoundsUntrustedFlowData(t *testing.T) {
	output := newMCPOutputBuilder(mcpUntrustedFlowDataNotice)
	output.Append(strings.Repeat("界", maxMCPToolResultBytes))
	got := output.String()

	if len(got) > maxMCPToolResultBytes {
		t.Fatalf("output length = %d, limit = %d", len(got), maxMCPToolResultBytes)
	}
	if !utf8.ValidString(got) {
		t.Fatalf("output is not valid UTF-8: %q", got)
	}
	if !strings.HasPrefix(got, mcpUntrustedFlowDataNotice) {
		t.Fatalf("output is missing the untrusted-data notice: %q", got[:min(len(got), 200)])
	}
	if !strings.Contains(got, mcpToolOutputTruncatedNote) {
		t.Fatalf("output is missing the truncation note: %q", got[len(got)-min(len(got), 200):])
	}
}

func TestTruncateMCPTextPreservesUTF8(t *testing.T) {
	if got := truncateMCPText("A界B", 2); got != "A" {
		t.Fatalf("truncated text = %q, want A", got)
	}
	if got := truncateMCPTextWithEllipsis("界界界", 7); got != "界..." {
		t.Fatalf("ellipsis text = %q, want 界...", got)
	}
}

func TestMCPAuditLogExcludesSensitiveContent(t *testing.T) {
	logger := logrus.StandardLogger()
	originalOutput := logger.Out
	originalFormatter := logger.Formatter
	originalLevel := logger.GetLevel()
	var output bytes.Buffer
	logger.SetOutput(&output)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	t.Cleanup(func() {
		logger.SetOutput(originalOutput)
		logger.SetFormatter(originalFormatter)
		logger.SetLevel(originalLevel)
	})

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "submit_flow_input",
			Arguments: map[string]any{
				"input": "sensitive flow input",
			},
		},
	}
	ctx := context.WithValue(context.Background(), mcpAuthModeContextKey{}, mcpAuthModeBearer)
	ctx = context.WithValue(ctx, mcpClientIPContextKey{}, "192.0.2.10")
	ctx = context.WithValue(ctx, mcpOriginHashContextKey{}, mcpIdentifierHash("https://console.example.com"))
	handler := auditToolCall(func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("sensitive tool result"), nil
	})

	if _, err := handler(ctx, request); err != nil {
		t.Fatalf("audit handler returned error: %v", err)
	}

	entry := output.String()
	for _, sensitiveValue := range []string{"sensitive flow input", "sensitive tool result", "192.0.2.10", "https://console.example.com"} {
		if strings.Contains(entry, sensitiveValue) {
			t.Fatalf("audit log contains sensitive value %q: %s", sensitiveValue, entry)
		}
	}
	for _, expectedField := range []string{"\"tool\":\"submit_flow_input\"", "\"auth_mode\":\"bearer\"", "\"access\":\"write\"", "\"outcome\":\"success\"", "\"client_ip_hash\":\"" + mcpIdentifierHash("192.0.2.10") + "\"", "\"origin_hash\":\"" + mcpIdentifierHash("https://console.example.com") + "\""} {
		if !strings.Contains(entry, expectedField) {
			t.Fatalf("audit log is missing %s: %s", expectedField, entry)
		}
	}
}

func TestLimitMCPToolResultRejectsOversizedWirePayload(t *testing.T) {
	ctx := withMCPAuditState(context.Background())
	handler := limitMCPToolResult(func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText(strings.Repeat("\"\n", maxMCPToolResultBytes)), nil
	})

	result, err := handler(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Name: mcpToolGetFlowStatus}})
	if err != nil {
		t.Fatalf("limit handler returned error: %v", err)
	}
	if result == nil || !result.IsError {
		t.Fatalf("oversized result = %#v, want tool error", result)
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal bounded result: %v", err)
	}
	if len(encoded) > maxMCPToolResultBytes {
		t.Fatalf("bounded wire size = %d, limit = %d", len(encoded), maxMCPToolResultBytes)
	}
	if got := mcpAuditDecisionFromContext(ctx); got != mcpAuditDecisionOutputLimit {
		t.Fatalf("audit decision = %q, want %s", got, mcpAuditDecisionOutputLimit)
	}
}

func TestLimitMCPToolResultPreservesSmallPayload(t *testing.T) {
	want := mcp.NewToolResultText("small result")
	handler := limitMCPToolResult(func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return want, nil
	})

	got, err := handler(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Name: mcpToolGetFlowStatus}})
	if err != nil {
		t.Fatalf("limit handler returned error: %v", err)
	}
	if got != want {
		t.Fatalf("small result pointer changed: got %p, want %p", got, want)
	}
}

func TestFlowBridgeSharesSSESessionBetweenHandlers(t *testing.T) {
	bridge := NewFlowBridge("PentAGI", "1.0.0", nil, nil)
	mux := http.NewServeMux()
	mux.Handle("/mcp/sse", bridge.SSEHandler())
	mux.Handle("/mcp/message", bridge.MessageHandler())
	testServer := httptest.NewServer(mux)
	t.Cleanup(testServer.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	sseRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL+"/mcp/sse", nil)
	if err != nil {
		t.Fatalf("create SSE request: %v", err)
	}
	sseResponse, err := http.DefaultClient.Do(sseRequest)
	if err != nil {
		t.Fatalf("open SSE connection: %v", err)
	}
	t.Cleanup(func() { _ = sseResponse.Body.Close() })
	if sseResponse.StatusCode != http.StatusOK {
		t.Fatalf("SSE status = %d, want %d", sseResponse.StatusCode, http.StatusOK)
	}

	messageEndpoint := ""
	reader := bufio.NewReader(sseResponse.Body)
	for messageEndpoint == "" {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read SSE endpoint event: %v", err)
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			messageEndpoint = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
	}

	messageURL := messageEndpoint
	if strings.HasPrefix(messageURL, "/") {
		messageURL = testServer.URL + messageURL
	}
	payload := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":%q,"clientInfo":{"name":"test","version":"1.0.0"},"capabilities":{}}}`, mcp.LATEST_PROTOCOL_VERSION)
	messageRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, messageURL, strings.NewReader(payload))
	if err != nil {
		t.Fatalf("create message request: %v", err)
	}
	messageRequest.Header.Set("Content-Type", "application/json")
	messageResponse, err := http.DefaultClient.Do(messageRequest)
	if err != nil {
		t.Fatalf("send message request: %v", err)
	}
	t.Cleanup(func() { _ = messageResponse.Body.Close() })
	if messageResponse.StatusCode != http.StatusAccepted {
		t.Fatalf("message status = %d, want %d", messageResponse.StatusCode, http.StatusAccepted)
	}
}

func TestFlowBridgeSetsBrowserCORSForConfiguredOrigin(t *testing.T) {
	const origin = "https://console.example.com"

	bridge := NewFlowBridgeWithOrigins("PentAGI", "1.0.0", nil, nil, false, []string{origin})
	testServer := httptest.NewServer(MCPOriginMiddleware([]string{origin})(bridge.SSEHandler()))
	t.Cleanup(testServer.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL+"/mcp/sse", nil)
	if err != nil {
		t.Fatalf("create SSE request: %v", err)
	}
	request.Header.Set("Origin", origin)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("open SSE connection: %v", err)
	}
	t.Cleanup(func() { _ = response.Body.Close() })

	if response.StatusCode != http.StatusOK {
		t.Fatalf("SSE status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if got := response.Header.Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, origin)
	}
}

func TestFlowBridgeCORSPreflightPassesBeforeBearerAuthentication(t *testing.T) {
	const origin = "https://console.example.com"

	bridge := NewFlowBridgeWithOrigins("PentAGI", "1.0.0", nil, nil, false, []string{origin})
	handler := MCPOriginMiddleware([]string{origin})(
		APIKeyMiddleware("mcp-test-key", false)(bridge.MessageHandler()),
	)

	preflight := httptest.NewRequest(http.MethodOptions, "/mcp/message", nil)
	preflight.Header.Set("Origin", origin)
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflight.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
	preflightRecorder := httptest.NewRecorder()
	handler.ServeHTTP(preflightRecorder, preflight)

	if preflightRecorder.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want %d", preflightRecorder.Code, http.StatusNoContent)
	}
	if got := preflightRecorder.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("preflight allow origin = %q, want %q", got, origin)
	}

	request := httptest.NewRequest(http.MethodPost, "/mcp/message", strings.NewReader(`{"jsonrpc":"2.0"}`))
	request.Header.Set("Origin", origin)
	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = -1
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated message status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func listToolNames(t *testing.T, bridge *FlowBridge) map[string]bool {
	t.Helper()

	request, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	if err != nil {
		t.Fatalf("marshal tools/list request: %v", err)
	}

	response := bridge.GetServer().HandleMessage(context.Background(), request)
	jsonResponse, ok := response.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("tools/list response = %T, want mcp.JSONRPCResponse", response)
	}
	result, ok := jsonResponse.Result.(mcp.ListToolsResult)
	if !ok {
		t.Fatalf("tools/list result = %T, want mcp.ListToolsResult", jsonResponse.Result)
	}

	names := make(map[string]bool, len(result.Tools))
	for _, tool := range result.Tools {
		names[tool.Name] = true
	}
	return names
}
