package config

import (
	"go.uber.org/fx"
)

var SignUpAPIConfigModule = fx.Options(
	fx.Provide(LoadDeployment("")),
	// fx.Provide(LoadDatabase("DATABASE")),
)
