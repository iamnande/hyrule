package app

import (
	"time"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/app"
)

func Initialize() []fx.Option {
	opts := make([]fx.Option, 0)
	opts = append(opts,
		fx.StartTimeout(5*time.Second),
		fx.StopTimeout(5*time.Second),
		// fx.Invoke(api.NewRESTServer),
		app.StartupModule,
		app.ShutdownModule,
	)
	return opts
}
