package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/cmd/pings/app"
	"github.com/iamnande/hyrule/go/internal/lib/config"
	"github.com/iamnande/hyrule/go/internal/lib/runtime"
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
)

func main() {
	fx.New(
		runtime.NewModule(app.Name),
		config.BaseModule,
		fx.Provide(config.LoadDatabase()),
		fx.Provide(domain.LoadConfig),
		app.Module,
	).Run()
}
