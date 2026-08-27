package config

import (
	"go.uber.org/fx"
)

var BaseModule = fx.Options(
	fx.Provide(LoadDeployment()),
	fx.Provide(LoadTracing()),
	fx.Provide(LoadHTTPServer()),
)
