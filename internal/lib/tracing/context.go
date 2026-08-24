package tracing

import (
	"context"

	"github.com/getsentry/sentry-go"
)

// CaptureException reports err against the current request's hub, if any is
// on ctx, falling back to the global hub otherwise.
func CaptureException(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
		return
	}
	sentry.CaptureException(err)
}

// SetTag annotates the current span in ctx, if any. A no-op otherwise.
func SetTag(ctx context.Context, key, value string) {
	if span := sentry.SpanFromContext(ctx); span != nil {
		span.SetTag(key, value)
	}
}

// TraceID returns the current trace ID from ctx. ok is false if there is no
// active span - callers must handle that case, not assume one always exists.
func TraceID(ctx context.Context) (id string, ok bool) {
	span := sentry.SpanFromContext(ctx)
	if span == nil {
		return "", false
	}
	return span.TraceID.String(), true
}

// SpanID returns the current span ID from ctx. ok is false if there is no
// active span.
func SpanID(ctx context.Context) (id string, ok bool) {
	span := sentry.SpanFromContext(ctx)
	if span == nil {
		return "", false
	}
	return span.SpanID.String(), true
}
