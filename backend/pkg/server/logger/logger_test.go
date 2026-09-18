package logger

import (
	"net/url"
	"testing"
)

func TestSanitizedRequestURI(t *testing.T) {
	testCases := []struct {
		name       string
		requestURL *url.URL
		want       string
	}{
		{
			name:       "path without query",
			requestURL: &url.URL{Path: "/api/v1/info"},
			want:       "/api/v1/info",
		},
		{
			name:       "redacts MCP session and API credentials",
			requestURL: &url.URL{Path: "/mcp/message", RawQuery: "sessionId=session-secret&api_key=key-secret&detail=tasks"},
			want:       "/mcp/message?api_key=%5BREDACTED%5D&detail=tasks&sessionId=%5BREDACTED%5D",
		},
		{
			name:       "redacts OAuth one-time values case insensitively",
			requestURL: &url.URL{Path: "/auth/callback", RawQuery: "CODE=code-secret&State=state-secret&next=%2Fdashboard"},
			want:       "/auth/callback?CODE=%5BREDACTED%5D&State=%5BREDACTED%5D&next=%2Fdashboard",
		},
		{
			name:       "does not emit malformed query data",
			requestURL: &url.URL{Path: "/mcp/message", RawQuery: "sessionId=session-secret%ZZ"},
			want:       "/mcp/message?query=REDACTED",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := sanitizedRequestURI(testCase.requestURL); got != testCase.want {
				t.Fatalf("sanitizedRequestURI() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestSensitiveQueryKey(t *testing.T) {
	for _, key := range []string{"sessionId", "session_id", "access-token", "API_KEY", "state", "signature"} {
		if !sensitiveQueryKey(key) {
			t.Errorf("sensitiveQueryKey(%q) = false, want true", key)
		}
	}
	for _, key := range []string{"flow_id", "detail", "page", "limit"} {
		if sensitiveQueryKey(key) {
			t.Errorf("sensitiveQueryKey(%q) = true, want false", key)
		}
	}
}
