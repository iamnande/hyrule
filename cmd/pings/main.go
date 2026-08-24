package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/cmd/pings/app"
	"github.com/iamnande/hyrule/internal/lib/config"
	"github.com/iamnande/hyrule/internal/lib/runtime"
	"github.com/iamnande/hyrule/internal/svc/pings/domain"
)

func main() {
	fx.New(
		runtime.NewModule(app.Name),
		config.PingsModule,
		fx.Provide(domain.LoadConfig),
		app.Module,
	).Run()
}
