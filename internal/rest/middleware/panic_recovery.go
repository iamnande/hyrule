package middleware

import (
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"

	"github.com/iamnande/hyrule/internal/services/logging"
)

func ResponsePanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if err := recover(); err != nil {
				logger := logging.FromContext(ctx)
				logger.Error("panic recovered", slog.Any("error", err))
				trace := sentry.GetHubFromContext(ctx)
				if trace != nil {
					trace.CaptureException(err.(error))
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
