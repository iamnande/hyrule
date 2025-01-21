package config

import (
	"go.uber.org/fx"
)

var SignUpAPIConfigModule = fx.Options(
	fx.Provide(LoadDeployment()),
	fx.Provide(LoadTracing()),
	fx.Provide(LoadDatabase()),
	fx.Provide(LoadJWT()),
)
