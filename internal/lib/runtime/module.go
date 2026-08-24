package runtime

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/rest"
	"github.com/iamnande/hyrule/internal/lib/version"
)

var HTTPModule = fx.Provide(
	NewHealthAPI,
	NewRouter,
)

func NewModule(serviceName string) fx.Option {
	return fx.Options(
		fx.Supply(version.NewServiceInfo(serviceName)),
		fx.StartTimeout(StartTimeout),
		fx.StopTimeout(StopTimeout),
		fx.Module("runtime",
			LoggingModule,
			TracingModule,
			StartupModule,
			ShutdownModule,
			HTTPModule,
			fx.Invoke(rest.NewServer),
		),
	)
}
