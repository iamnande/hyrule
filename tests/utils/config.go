package utils

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
)

func TestConfigs() []fx.Option {
	return []fx.Option{
		fx.Decorate(func() config.Deployment {
			return config.Deployment{
				Region:      config.PrimaryRegion,
				Environment: config.LocalEnvironment,
			}
		}),
	}
}
