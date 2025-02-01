package config

import (
	"go.uber.org/fx"
)

var RegistrationAPIModule = fx.Options(
	fx.Provide(LoadDeployment()),
	fx.Provide(LoadTracing()),
	fx.Provide(LoadDatabase()),
	fx.Provide(LoadJWT()),
	fx.Provide(LoadEmail()),
)
