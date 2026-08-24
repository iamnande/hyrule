package tracing

import (
	"context"
	goruntime "runtime"

	"github.com/getsentry/sentry-go"
)

// Start begins a span named after its caller, so call sites never type a
// name: ctx, done := tracing.Start(ctx); defer done()
func Start(ctx context.Context) (context.Context, func()) {
	name := "unknown"
	if pc, _, _, ok := goruntime.Caller(1); ok {
		if fn := goruntime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}
	}
	return StartNamed(ctx, name)
}

// StartNamed begins a span with an explicit name - for the rare case where
// the name is data-dependent (e.g. one span per dependency check) and can't
// be derived from the calling function the way Start's can.
func StartNamed(ctx context.Context, name string) (context.Context, func()) {
	span := sentry.StartSpan(ctx, name)
	return span.Context(), span.Finish
}
