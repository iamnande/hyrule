package app

import (
	"context"
	"log/slog"

	"go.uber.org/fx"
)

var StartupModule = fx.Module("startup",
	fx.Invoke(func(lifecycle fx.Lifecycle, logger *slog.Logger) error {
		lifecycle.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				logger.Info("initializing service")
				return nil
			},
		})
		return nil
	}),
)

// TODO: shared startup checks
