package app

import (
	"go.uber.org/fx"

	healthAPI "github.com/iamnande/hyrule/internal/apis/signup-api/health"
	v1SignUpAPIRouter "github.com/iamnande/hyrule/internal/apis/signup-api/v1"
	v1SignUpAPI "github.com/iamnande/hyrule/internal/apis/signup-api/v1/signup"
	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/database"
	"github.com/iamnande/hyrule/internal/services/signup"
)

func Build() []fx.Option {
	return []fx.Option{
		config.SignUpAPIConfigModule,

		// services
		fx.Provide(
			fx.Annotate(database.NewDatabaseClient,
				fx.As(new(healthAPI.DatabaseClient)),
			),
			fx.Annotate(signup.NewService, fx.As(new(v1SignUpAPI.SignUpService))),
		),

		// runtime
		fx.Provide(
			healthAPI.NewHealthAPI,
			v1SignUpAPI.NewSignUpAPI,
			v1SignUpAPIRouter.NewSignUpAPIRouter,
		),
	}
}
