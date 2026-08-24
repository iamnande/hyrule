package config

import (
	"go.uber.org/fx"
)

var PingsModule = fx.Options(
	fx.Provide(LoadDeployment()),
	fx.Provide(LoadTracing()),
	fx.Provide(LoadHTTPServer()),
	fx.Provide(LoadDatabase()),
)
