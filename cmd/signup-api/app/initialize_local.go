//go:build local

package app

import (
	"time"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/app"
	"github.com/iamnande/hyrule/internal/rest"
)

func Initialize() []fx.Option {
	opts := make([]fx.Option, 0)
	opts = append(opts,
		fx.StartTimeout(5*time.Second),
		fx.StopTimeout(5*time.Second),
		fx.Invoke(rest.NewServer),
		app.StartupModule,
		app.ShutdownModule,
		// TODO: local db override
	)
	return opts
}
