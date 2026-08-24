package middleware

import (
	"net/http"
	"slices"

	"github.com/iamnande/hyrule/internal/lib/tracing"
)

var (
	tracerIgnoredURLPaths = []string{
		"/discovery",
		"/healthz",
		"/startupz",
		"/livez",
		"/readyz",
	}
)

func RequestTracer(next http.Handler) http.Handler {
	traced := tracing.Middleware(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(tracerIgnoredURLPaths, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		traced.ServeHTTP(w, r)
	})
}
