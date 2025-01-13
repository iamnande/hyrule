package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/iamnande/hyrule/internal/services/logging"
)

func ResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrappedWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(wrappedWriter, r.WithContext(r.Context()))
		logger := logging.FromContext(r.Context())
		// TODO: share attributes and log
		// TODO: wire up tracing
		if wrappedWriter.Status() >= 400 {
			logger.Error(fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, wrappedWriter.Status()),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status_code", wrappedWriter.Status()),
			)
			return
		}
		logger.Info(fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, wrappedWriter.Status()),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status_code", wrappedWriter.Status()),
		)
	})
}
