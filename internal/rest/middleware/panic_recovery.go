package middleware

import (
	"log/slog"
	"net/http"

	"github.com/iamnande/hyrule/internal/services/logging"
)

func ResponsePanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := logging.FromContext(r.Context())
				logger.Error("panic recovered", slog.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
				// TODO: rest - marshal json
			}
		}()
		next.ServeHTTP(w, r)
	})
}
