package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/cmd/iam-jwks/app"
	"github.com/iamnande/hyrule/internal/lib/config"
	"github.com/iamnande/hyrule/internal/lib/runtime"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/repository"
)

func main() {
	fileCfg, err := repository.LoadFileConfig()
	if err != nil {
		panic(err)
	}

	fx.New(
		runtime.NewModule(app.Name),
		config.BaseModule,
		fx.Provide(config.LoadDatabase()),
		app.Module(fileCfg),
	).Run()
}
