package middleware

import (
	"log/slog"
	"net/http"

	"github.com/iamnande/hyrule/internal/services/logging"
)

func RequestLogger(logger *slog.Logger) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(logging.WithContext(r.Context(), logger)))
		})
	}
}
