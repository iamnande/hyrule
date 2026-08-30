package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/cmd/iam-jwks/app"
	"github.com/iamnande/hyrule/go/internal/lib/config"
	"github.com/iamnande/hyrule/go/internal/lib/runtime"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/repository"
)

func main() {
	fx.New(
		runtime.NewModule(app.Name),
		config.BaseModule,
		fx.Provide(repository.LoadEnvConfig),
		app.Module,
	).Run()
}
