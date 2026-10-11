package mcpbridge

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"
)

type mcpAuthModeContextKey struct{}
type mcpOriginValidatedContextKey struct{}
type mcpOriginHashContextKey struct{}
type mcpAccessScopeContextKey struct{}

const (
	mcpAuthModeAnonymous            = "anonymous"
	mcpAuthModeBearer               = "bearer"
	mcpAuthModeUnknown              = "unknown"
	mcpAccessScopeUnknown           = "unknown"
	mcpAccessScopeRead              = "read"
	mcpAccessScopeWrite             = "write"
	DefaultMCPMaxRequestBytes int64 = 1 << 20
)

// APIKeyMiddleware returns a middleware that requires a valid MCP API key.
// Anonymous access must be explicitly enabled when no key is configured.
func APIKeyMiddleware(expectedKey string, allowAnonymous bool) func(http.Handler) http.Handler {
	return APIKeyMiddlewareWithWriteKey(expectedKey, "", allowAnonymous)
}

// APIKeyMiddlewareWithWriteKey authenticates read access and optionally assigns
// a separate write scope when a write key is configured.
func APIKeyMiddlewareWithWriteKey(expectedKey, writeKey string, allowAnonymous bool) func(http.Handler) http.Handler {
	expectedKey = strings.TrimSpace(expectedKey)
	writeKey = strings.TrimSpace(writeKey)

	return func(next http.Handler) http.Handler {
		if expectedKey == "" {
			if allowAnonymous {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, withMCPAuthContext(r, mcpAuthModeAnonymous, mcpAccessScopeRead))
				})
			}
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "MCP API key is not configured", http.StatusServiceUnavailable)
			})
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isMCPPreflight(r) && mcpOriginWasValidated(r) {
				next.ServeHTTP(w, r)
				return
			}
			token := bearerToken(r.Header.Get("Authorization"))
			scope := ""
			if writeKey != "" && mcpSecretEqual(token, writeKey) {
				scope = mcpAccessScopeWrite
			} else if mcpSecretEqual(token, expectedKey) {
				if writeKey == "" {
					scope = mcpAccessScopeWrite
				} else {
					scope = mcpAccessScopeRead
				}
			}
			if scope != "" {
				authenticatedRequest := withMCPAuthContext(r, mcpAuthModeBearer, scope)
				next.ServeHTTP(w, withMCPPrincipalHash(authenticatedRequest, token))
				return
			}
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}
}

func mcpSecretEqual(provided, expected string) bool {
	providedDigest := sha256.Sum256([]byte(provided))
	expectedDigest := sha256.Sum256([]byte(expected))
	return subtle.ConstantTimeCompare(providedDigest[:], expectedDigest[:]) == 1
}

func isMCPPreflight(r *http.Request) bool {
	return r != nil && r.Method == http.MethodOptions &&
		r.Header.Get("Origin") != "" &&
		r.Header.Get("Access-Control-Request-Method") != ""
}

func mcpOriginWasValidated(r *http.Request) bool {
	if r == nil {
		return false
	}
	validated, _ := r.Context().Value(mcpOriginValidatedContextKey{}).(bool)
	return validated
}

// MCPOriginMiddleware rejects browser requests whose Origin is not explicitly allowed.
// Requests without an Origin header remain compatible with desktop MCP clients.
func MCPOriginMiddleware(origins []string) func(http.Handler) http.Handler {
	allowedOrigins := normalizedMCPOrigins(origins)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				normalizedOrigin, valid := normalizeMCPOrigin(origin)
				if !valid || !mcpNormalizedOriginAllowed(normalizedOrigin, allowedOrigins) {
					w.Header().Add("Vary", "Origin")
					http.Error(w, "MCP origin is not allowed", http.StatusForbidden)
					return
				}
				ctx := context.WithValue(r.Context(), mcpOriginValidatedContextKey{}, true)
				ctx = context.WithValue(ctx, mcpOriginHashContextKey{}, mcpIdentifierHash(normalizedOrigin))
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func MCPRequestBodyLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = DefaultMCPMaxRequestBytes
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || (r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch) {
				next.ServeHTTP(w, r)
				return
			}
			if r.ContentLength > maxBytes {
				http.Error(w, "MCP request body is too large", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func withMCPAuthContext(r *http.Request, mode, scope string) *http.Request {
	ctx := context.WithValue(r.Context(), mcpAuthModeContextKey{}, mode)
	return r.WithContext(context.WithValue(ctx, mcpAccessScopeContextKey{}, scope))
}

func withMCPAuthMode(r *http.Request, mode string) *http.Request {
	return withMCPAuthContext(r, mode, mcpAccessScopeRead)
}

func mcpAuthModeFromContext(ctx context.Context) string {
	mode, ok := ctx.Value(mcpAuthModeContextKey{}).(string)
	if !ok || mode == "" {
		return mcpAuthModeUnknown
	}
	return mode
}

func mcpAccessScopeFromContext(ctx context.Context) string {
	scope, ok := ctx.Value(mcpAccessScopeContextKey{}).(string)
	if !ok || scope == "" {
		return mcpAccessScopeUnknown
	}
	return scope
}

func bearerToken(authorization string) string {
	scheme, token, found := strings.Cut(authorization, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func normalizedMCPOrigins(origins []string) []string {
	allowedOrigins := make([]string, 0, len(origins))
	seen := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "*" {
			return []string{"*"}
		}
		normalizedOrigin, ok := normalizeMCPOrigin(origin)
		if !ok {
			continue
		}
		if _, exists := seen[normalizedOrigin]; exists {
			continue
		}
		seen[normalizedOrigin] = struct{}{}
		allowedOrigins = append(allowedOrigins, normalizedOrigin)
	}
	return allowedOrigins
}

func mcpOriginAllowed(origin string, allowedOrigins []string) bool {
	normalizedOrigin, ok := normalizeMCPOrigin(origin)
	if !ok {
		return false
	}
	return mcpNormalizedOriginAllowed(normalizedOrigin, allowedOrigins)
}

func mcpNormalizedOriginAllowed(normalizedOrigin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == normalizedOrigin {
			return true
		}
	}
	return false
}

func mcpOriginHashFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	originHash, _ := ctx.Value(mcpOriginHashContextKey{}).(string)
	return strings.TrimSpace(originHash)
}

func normalizeMCPOrigin(origin string) (string, bool) {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(origin))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return "", false
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", false
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host), true
}
