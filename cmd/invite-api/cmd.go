package adminapi

import (
	"log/slog"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/iamnande/hyrule/cmd/invite-api/app"
	modules "github.com/iamnande/hyrule/internal/app"
)

func Run() {
	opts := []fx.Option{}
	opts = append(opts, []fx.Option{
		modules.LoggingModule,
		modules.TracingModule,
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return &fxevent.SlogLogger{Logger: logger}
		}),
	}...)
	opts = append(opts, app.Build()...)
	opts = append(opts, app.Initialize()...)
	fx.New(opts...).Run()
}
