package logging

import (
	"context"
	"log/slog"
)

type contextKey struct{}

var (
	ContextKey = contextKey{}
)

func FromContext(ctx context.Context) *slog.Logger {
	logger, exists := ctx.Value(ContextKey).(*slog.Logger)
	if !exists {
		return slog.Default()
	}
	return logger
}

func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ContextKey, logger)
}
