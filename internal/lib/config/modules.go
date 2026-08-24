package config

import (
	"go.uber.org/fx"
)

var PingsModule = fx.Options(
	fx.Provide(LoadDeployment()),
	fx.Provide(LoadTracing()),
	fx.Provide(LoadHTTPServer()),
	// TODO: LoadDatabase() joins here once the postgres data layer lands.
)
