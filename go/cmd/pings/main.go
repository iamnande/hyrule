package main

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/cmd/pings/app"
	"github.com/iamnande/hyrule/go/internal/lib/config"
	"github.com/iamnande/hyrule/go/internal/lib/runtime"
)

func main() {
	fx.New(
		runtime.NewModule(app.Name),
		config.BaseModule,
		app.Module,
	).Run()
}
