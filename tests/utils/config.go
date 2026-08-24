package utils

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/config"
	"github.com/iamnande/hyrule/internal/svc/pings/domain"
)

func TestConfigs() []fx.Option {
	return []fx.Option{
		fx.Provide(config.LoadTracing()),
		fx.Supply(config.Deployment{
			Region:      config.PrimaryRegion,
			Environment: config.LocalEnvironment,
		}),
		fx.Supply(config.HTTPServer{Addr: ":0"}),
		fx.Provide(config.LoadDatabase()),
		fx.Provide(domain.LoadConfig),
	}
}
