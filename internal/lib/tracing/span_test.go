package tracing_test

import (
	"context"
	"testing"

	"github.com/getsentry/sentry-go"

	"github.com/iamnande/hyrule/internal/lib/tracing"
)

func TestStart_NamesSpanAfterCaller(t *testing.T) {
	ctx, done := tracing.Start(context.Background())
	defer done()

	span := sentry.SpanFromContext(ctx)
	if span == nil {
		t.Fatal("expected a span in the returned context")
	}
	const want = "github.com/iamnande/hyrule/internal/lib/tracing_test.TestStart_NamesSpanAfterCaller"
	if span.Op != want {
		t.Errorf("span op = %q, want %q", span.Op, want)
	}
}
