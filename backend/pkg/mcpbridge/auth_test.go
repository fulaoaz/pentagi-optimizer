package mcpbridge

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIKeyMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		expectedKey    string
		allowAnonymous bool
		authorization  string
		query          string
		wantStatus     int
		wantChallenge  string
		wantAuthMode   string
	}{
		{
			name:          "accepts bearer key",
			expectedKey:   "mcp-test-key",
			authorization: "Bearer mcp-test-key",
			wantStatus:    http.StatusNoContent,
			wantAuthMode:  mcpAuthModeBearer,
		},
		{
			name:          "accepts case insensitive bearer scheme",
			expectedKey:   "mcp-test-key",
			authorization: "bearer mcp-test-key",
			wantStatus:    http.StatusNoContent,
			wantAuthMode:  mcpAuthModeBearer,
		},
		{
			name:          "rejects missing bearer key",
			expectedKey:   "mcp-test-key",
			wantStatus:    http.StatusUnauthorized,
			wantChallenge: "Bearer",
		},
		{
			name:          "rejects query parameter key",
			expectedKey:   "mcp-test-key",
			query:         "?api_key=mcp-test-key",
			wantStatus:    http.StatusUnauthorized,
			wantChallenge: "Bearer",
		},
		{
			name:       "rejects unconfigured protected bridge",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "allows explicitly anonymous bridge",
			allowAnonymous: true,
			wantStatus:     http.StatusNoContent,
			wantAuthMode:   mcpAuthModeAnonymous,
		},
		{
			name:           "configured key still wins over anonymous mode",
			expectedKey:    "mcp-test-key",
			allowAnonymous: true,
			wantStatus:     http.StatusUnauthorized,
			wantChallenge:  "Bearer",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			authMode := ""
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authMode = mcpAuthModeFromContext(r.Context())
				w.WriteHeader(http.StatusNoContent)
			})
			handler := APIKeyMiddleware(testCase.expectedKey, testCase.allowAnonymous)(next)

			request := httptest.NewRequest(http.MethodGet, "/mcp"+testCase.query, nil)
			if testCase.authorization != "" {
				request.Header.Set("Authorization", testCase.authorization)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.wantStatus)
			}
			if got := recorder.Header().Get("WWW-Authenticate"); got != testCase.wantChallenge {
				t.Fatalf("WWW-Authenticate = %q, want %q", got, testCase.wantChallenge)
			}
			if authMode != testCase.wantAuthMode {
				t.Fatalf("auth mode = %q, want %q", authMode, testCase.wantAuthMode)
			}
		})
	}
}

func TestAPIKeyMiddlewareAllowsCORSPreflightWithoutCredentials(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})
	const origin = "https://console.example.com"
	handler := MCPOriginMiddleware([]string{origin})(APIKeyMiddleware("mcp-test-key", false)(next))

	request := httptest.NewRequest(http.MethodOptions, "/mcp/message", nil)
	request.Header.Set("Origin", origin)
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if !nextCalled {
		t.Fatal("preflight was not passed to the CORS handler")
	}
}

func TestAPIKeyMiddlewareDoesNotTreatOrdinaryOptionsAsPreflight(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})
	handler := APIKeyMiddleware("mcp-test-key", false)(next)

	request := httptest.NewRequest(http.MethodOptions, "/mcp/message", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Fatal("ordinary OPTIONS request bypassed authentication")
	}
}

func TestAPIKeyMiddlewareRequiresOriginValidationForPreflight(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})
	handler := APIKeyMiddleware("mcp-test-key", false)(next)

	request := httptest.NewRequest(http.MethodOptions, "/mcp/message", nil)
	request.Header.Set("Origin", "https://console.example.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if nextCalled {
		t.Fatal("preflight without origin validation bypassed authentication")
	}
}

func TestAPIKeyMiddlewareWithWriteKeyAssignsScopes(t *testing.T) {
	testCases := []struct {
		name          string
		authorization string
		wantStatus    int
		wantScope     string
	}{
		{name: "read key", authorization: "Bearer read-key", wantStatus: http.StatusNoContent, wantScope: mcpAccessScopeRead},
		{name: "write key", authorization: "Bearer write-key", wantStatus: http.StatusNoContent, wantScope: mcpAccessScopeWrite},
		{name: "invalid key", authorization: "Bearer wrong-key", wantStatus: http.StatusUnauthorized},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			scope := ""
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				scope = mcpAccessScopeFromContext(r.Context())
				w.WriteHeader(http.StatusNoContent)
			})
			handler := APIKeyMiddlewareWithWriteKey("read-key", "write-key", false)(next)
			request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
			request.Header.Set("Authorization", testCase.authorization)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.wantStatus)
			}
			if scope != testCase.wantScope {
				t.Fatalf("scope = %q, want %q", scope, testCase.wantScope)
			}
		})
	}
}

func TestMCPSecretEqual(t *testing.T) {
	testCases := []struct {
		name     string
		provided string
		expected string
		want     bool
	}{
		{name: "matching values", provided: "mcp-test-key", expected: "mcp-test-key", want: true},
		{name: "different values", provided: "mcp-test-key", expected: "mcp-other-key", want: false},
		{name: "different lengths", provided: "x", expected: "mcp-test-key", want: false},
		{name: "empty values", want: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := mcpSecretEqual(testCase.provided, testCase.expected); got != testCase.want {
				t.Fatalf("mcpSecretEqual() = %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestMCPOriginMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		allowedOrigins []string
		origin         string
		wantStatus     int
		wantNext       bool
	}{
		{
			name:       "allows desktop client without origin",
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
		{
			name:       "rejects browser origin by default",
			origin:     "https://console.example.com",
			wantStatus: http.StatusForbidden,
		},
		{
			name:           "allows configured browser origin",
			allowedOrigins: []string{"https://console.example.com"},
			origin:         "https://console.example.com",
			wantStatus:     http.StatusNoContent,
			wantNext:       true,
		},
		{
			name:           "normalizes configured origin",
			allowedOrigins: []string{"HTTPS://Console.Example.com/"},
			origin:         "https://console.example.com",
			wantStatus:     http.StatusNoContent,
			wantNext:       true,
		},
		{
			name:           "allows any valid origin only when explicit",
			allowedOrigins: []string{"*"},
			origin:         "https://console.example.com",
			wantStatus:     http.StatusNoContent,
			wantNext:       true,
		},
		{
			name:           "rejects opaque origin even with wildcard",
			allowedOrigins: []string{"*"},
			origin:         "null",
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "rejects origin outside configured allowlist",
			allowedOrigins: []string{"https://console.example.com"},
			origin:         "https://attacker.example",
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "rejects non HTTP origin",
			allowedOrigins: []string{"ftp://console.example.com"},
			origin:         "ftp://console.example.com",
			wantStatus:     http.StatusForbidden,
		},
		{
			name:           "rejects origin with a path",
			allowedOrigins: []string{"https://console.example.com/app"},
			origin:         "https://console.example.com/app",
			wantStatus:     http.StatusForbidden,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusNoContent)
			})
			handler := MCPOriginMiddleware(testCase.allowedOrigins)(next)
			request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
			if testCase.origin != "" {
				request.Header.Set("Origin", testCase.origin)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.wantStatus)
			}
			if nextCalled != testCase.wantNext {
				t.Fatalf("next called = %t, want %t", nextCalled, testCase.wantNext)
			}
		})
	}
}

func TestMCPOriginMiddlewareAddsNormalizedOriginHash(t *testing.T) {
	const origin = "https://console.example.com"
	originHash := ""
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originHash = mcpOriginHashFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})
	handler := MCPOriginMiddleware([]string{"HTTPS://Console.Example.com/"})(next)
	request := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	request.Header.Set("Origin", origin)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if want := mcpIdentifierHash(origin); originHash != want {
		t.Fatalf("origin hash = %q, want %q", originHash, want)
	}
	if strings.Contains(originHash, origin) {
		t.Fatalf("origin hash exposes the raw origin: %q", originHash)
	}
}

func TestMCPRequestBodyLimitMiddleware(t *testing.T) {
	testCases := []struct {
		name          string
		method        string
		body          string
		maxBytes      int64
		wantStatus    int
		wantNext      bool
		wantReadError bool
	}{
		{
			name:       "allows body under limit",
			method:     http.MethodPost,
			body:       "small body",
			maxBytes:   64,
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
		{
			name:       "rejects oversized declared body",
			method:     http.MethodPost,
			body:       strings.Repeat("x", 65),
			maxBytes:   64,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:          "limits streamed body",
			method:        http.MethodPost,
			body:          strings.Repeat("x", 65),
			maxBytes:      64,
			wantStatus:    http.StatusBadRequest,
			wantNext:      true,
			wantReadError: true,
		},
		{
			name:       "does not limit GET stream",
			method:     http.MethodGet,
			body:       strings.Repeat("x", 65),
			maxBytes:   64,
			wantStatus: http.StatusNoContent,
			wantNext:   true,
		},
		{
			name:       "uses default when configured limit is nonpositive",
			method:     http.MethodPost,
			body:       strings.Repeat("x", int(DefaultMCPMaxRequestBytes)+1),
			maxBytes:   0,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if testCase.wantReadError {
					if _, err := io.ReadAll(r.Body); err == nil {
						t.Error("body read error = nil, want size limit error")
					} else {
						http.Error(w, "request body too large", http.StatusBadRequest)
						return
					}
				}
				w.WriteHeader(http.StatusNoContent)
			})
			handler := MCPRequestBodyLimitMiddleware(testCase.maxBytes)(next)
			request := httptest.NewRequest(testCase.method, "/mcp/message", strings.NewReader(testCase.body))
			if testCase.name == "limits streamed body" {
				request.ContentLength = -1
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != testCase.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, testCase.wantStatus)
			}
			if nextCalled != testCase.wantNext {
				t.Fatalf("next called = %t, want %t", nextCalled, testCase.wantNext)
			}
		})
	}
}
