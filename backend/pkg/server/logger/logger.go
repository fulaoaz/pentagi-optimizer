package logger

import (
	"context"
	"net/url"
	"strings"
	"time"

	obs "pentagi/pkg/observability"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// FromContext is function to get logrus Entry with context
func FromContext(c *gin.Context) *logrus.Entry {
	return logrus.WithContext(c.Request.Context())
}

func TraceEnabled() bool {
	return logrus.IsLevelEnabled(logrus.TraceLevel)
}

func WithGinLogger(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		uri := sanitizedRequestURI(c.Request.URL)

		entry := logrus.WithFields(logrus.Fields{
			"component":      "api",
			"net_peer_ip":    c.ClientIP(),
			"http_uri":       uri,
			"http_path":      c.Request.URL.Path,
			"http_host_name": c.Request.Host,
			"http_method":    c.Request.Method,
		})
		if c.FullPath() == "" {
			entry = entry.WithField("request", "proxy handled")
		} else {
			entry = entry.WithField("request", "api handled")
		}

		// serve the request to the next middleware
		c.Next()

		if len(c.Errors) > 0 {
			entry = entry.WithField("gin.errors", c.Errors.String())
		}

		entry = entry.WithFields(logrus.Fields{
			"duration":         time.Since(start).String(),
			"http_status_code": c.Writer.Status(),
			"http_resp_size":   c.Writer.Size(),
		}).WithContext(c.Request.Context())
		if requestID := c.Writer.Header().Get("X-Request-ID"); requestID != "" {
			entry = entry.WithField("request_id", requestID)
		}
		if c.Writer.Status() >= 400 {
			entry.Error("http request handled error")
		} else {
			entry.Debug("http request handled success")
		}
	}
}

const redactedQueryValue = "[REDACTED]"

// sanitizedRequestURI preserves useful query keys while keeping credentials,
// session identifiers, and one-time authorization values out of access logs.
func sanitizedRequestURI(requestURL *url.URL) string {
	if requestURL == nil {
		return ""
	}

	uri := requestURL.Path
	if uri == "" {
		uri = "/"
	}
	if requestURL.RawQuery == "" {
		return uri
	}

	query, err := url.ParseQuery(requestURL.RawQuery)
	if err != nil {
		return uri + "?query=REDACTED"
	}
	for key := range query {
		if sensitiveQueryKey(key) {
			query[key] = []string{redactedQueryValue}
		}
	}
	if encodedQuery := query.Encode(); encodedQuery != "" {
		return uri + "?" + encodedQuery
	}
	return uri
}

func sensitiveQueryKey(key string) bool {
	normalized := strings.ToLower(strings.NewReplacer("-", "", "_", "").Replace(key))
	switch normalized {
	case "apikey", "accesstoken", "refreshtoken", "idtoken", "token", "sessionid", "authorization", "password", "secret", "clientsecret", "code", "state", "signature", "sig":
		return true
	default:
		return false
	}
}

func WithGqlLogger(service string) func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	return func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		ctx, span := obs.Observer.NewSpan(ctx, obs.SpanKindServer, "graphql.handler")
		defer span.End()

		start := time.Now()
		entry := logrus.WithContext(ctx).WithField("component", service)

		res := next(ctx)

		op := graphql.GetOperationContext(ctx)
		if op != nil && op.Operation != nil {
			entry = entry.WithFields(logrus.Fields{
				"operation_name": op.OperationName,
				"operation_type": op.Operation.Operation,
			})
		}

		entry = entry.WithField("duration", time.Since(start).String())

		if res == nil {
			return res
		}

		if len(res.Errors) > 0 {
			entry = entry.WithField("gql.errors", res.Errors.Error())
			entry.Error("graphql request handled with errors")
		} else {
			entry.Debug("graphql request handled success")
		}

		return res
	}
}
