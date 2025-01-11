package config

import (
	"go.uber.org/fx"
)

var AdminAPIConfigModule = fx.Options(
	fx.Provide(LoadDeployment("")),
	// fx.Provide(LoadDatabase("DATABASE")),
)
