package tracing

import (
	"context"
	"fmt"
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

// StartNamed begins a span with an explicit static name - for the rare case
// where the name isn't data-dependent but still can't be the calling
// function's name (see StartNamedf for the data-dependent case).
func StartNamed(ctx context.Context, name string) (context.Context, func()) {
	span := sentry.StartSpan(ctx, name)
	return span.Context(), span.Finish
}

// StartNamedf is StartNamed with the name built via fmt.Sprintf - for one
// span per item in a loop (e.g. "dependencies.check.%s" per dependency).
func StartNamedf(ctx context.Context, format string, args ...any) (context.Context, func()) {
	return StartNamed(ctx, fmt.Sprintf(format, args...))
}
