package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/iamnande/hyrule/internal/services/logging"
)

const (
	OpName = "api.rest"
)

func ResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// before API
		var (
			ctx             = r.Context()
			start           = time.Now()
			logger          = logging.FromContext(ctx)
			wrappedWriter   = middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			transactionName = fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		)

		// start transaction
		span := sentry.StartTransaction(ctx, transactionName, sentry.WithOpName(OpName))
		defer span.Finish()

		// serve API
		next.ServeHTTP(wrappedWriter, r.WithContext(span.Context()))

		// after API
		message := fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, wrappedWriter.Status())
		attributes := []any{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status_code", wrappedWriter.Status()),
			slog.String("duration", time.Since(start).String()),
		}

		if span != nil {
			attributes = append(attributes,
				slog.String("trace_id", span.TraceID.String()),
				slog.String("span_id", span.SpanID.String()),
			)
		}

		if wrappedWriter.Status() >= 400 {

			logger.Error(message, attributes...)
			return
		}

		logger.Info(message, attributes...)
	})
}
