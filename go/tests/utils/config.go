package utils

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/config"
)

func TestConfigs() []fx.Option {
	return []fx.Option{
		fx.Provide(config.LoadTracing()),
		fx.Supply(config.Deployment{
			Region:      config.USEast2Region,
			Environment: config.LocalEnvironment,
		}),
		fx.Supply(config.HTTPServer{Addr: ":0"}),
		fx.Provide(config.LoadDatabase()),
	}
}
