package utils

import (
	"github.com/iamnande/hyrule/internal/config"
	"go.uber.org/fx"
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
