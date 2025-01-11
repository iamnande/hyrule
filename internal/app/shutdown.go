package app

import (
	"context"
	"time"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
)

var ShutdownModule = fx.Module("shutdown",
	fx.Invoke(func(lifecycle fx.Lifecycle, deployment config.Deployment) error {
		lifecycle.Append(fx.Hook{
			OnStop: func(_ context.Context) error {
				if deployment.Environment == config.LocalEnvironment {
					return nil
				}
				time.Sleep(5 * time.Second)
				return nil
			},
		})
		return nil
	}),
)
