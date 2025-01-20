package middleware

import (
	"net/http"
	"slices"

	sentryhttp "github.com/getsentry/sentry-go/http"
)

var (
	tracerIgnoredURLPaths = []string{
		"/health",
		"/discovery",
	}
)

func RequestTracer(next http.Handler) http.Handler {
	tracer := sentryhttp.New(sentryhttp.Options{})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(tracerIgnoredURLPaths, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		tracer.Handle(next).ServeHTTP(w, r)
	})
}
