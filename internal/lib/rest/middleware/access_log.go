package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/iamnande/hyrule/internal/lib/logging"
	"github.com/iamnande/hyrule/internal/lib/tracing"
)

func ResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx           = r.Context()
			start         = time.Now()
			logger        = logging.FromContext(ctx)
			wrappedWriter = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		)

		next.ServeHTTP(wrappedWriter, r)

		message := fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, wrappedWriter.Status())
		attributes := []any{
			slog.Group("request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			),
			slog.Group("response",
				slog.Int("status_code", wrappedWriter.Status()),
				slog.String("duration", time.Since(start).String()),
			),
		}

		if traceID, ok := tracing.TraceID(ctx); ok {
			spanID, _ := tracing.SpanID(ctx)
			attributes = append(attributes,
				slog.Group("tracing",
					slog.String("trace_id", traceID),
					slog.String("span_id", spanID),
				),
			)
		}

		if wrappedWriter.Status() >= 400 {
			logger.Error(message, attributes...)
			return
		}

		logger.Info(message, attributes...)
	})
}
