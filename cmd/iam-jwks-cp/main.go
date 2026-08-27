package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/cmd/iam-jwks-cp/app"
	"github.com/iamnande/hyrule/internal/lib/config"
	"github.com/iamnande/hyrule/internal/lib/runtime"
)

func main() {
	fx.New(
		runtime.NewModule(app.Name),
		config.BaseModule,
		fx.Provide(config.LoadDatabase()),
		app.Module,
	).Run()
}
