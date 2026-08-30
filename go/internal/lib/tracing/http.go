package tracing

import (
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
)

// Middleware wraps a handler with automatic request-level tracing - one
// span per request, attached to the request's context.
func Middleware(next http.Handler) http.Handler {
	return sentryhttp.New(sentryhttp.Options{}).Handle(next)
}
