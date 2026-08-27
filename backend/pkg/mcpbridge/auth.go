package mcpbridge

import (
	"crypto/subtle"
	"net/http"
)

// APIKeyMiddleware returns a middleware that requires a valid MCP API key.
// If expectedKey is empty, the middleware is a no-op (anonymous local mode).
// The key can be supplied via Authorization: Bearer <key> header or api_key query.
func APIKeyMiddleware(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if expectedKey == "" {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			authz := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if len(authz) > len(prefix) && authz[:len(prefix)] == prefix {
				token = authz[len(prefix):]
			}
			if token == "" {
				token = r.URL.Query().Get("api_key")
			}
			if subtle.ConstantTimeCompare([]byte(token), []byte(expectedKey)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "unauthorized: invalid MCP API key", http.StatusUnauthorized)
		})
	}
}
