package middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/iamnande/hyrule/go/internal/lib/logging"
	"github.com/iamnande/hyrule/go/internal/lib/tracing"
)

func ResponsePanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer func() {
			if recovered := recover(); recovered != nil {
				logger := logging.FromContext(ctx)
				logger.Error("panic recovered", slog.Any("error", recovered))
				err, ok := recovered.(error)
				if !ok {
					err = fmt.Errorf("panic: %v", recovered)
				}
				tracing.CaptureException(ctx, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
