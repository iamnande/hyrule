package app

import (
	"go.uber.org/fx"

	healthAPI "github.com/iamnande/hyrule/internal/apis/signup-api/health"
	v1SignUpAPIRouter "github.com/iamnande/hyrule/internal/apis/signup-api/v1"
	v1SignUpAPI "github.com/iamnande/hyrule/internal/apis/signup-api/v1/signup"
	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/services/signup"
)

func Build() []fx.Option {
	return []fx.Option{
		config.SignUpAPIConfigModule,

		// TODO: service layer
		fx.Provide(
			fx.Annotate(signup.NewService, fx.As(new(v1SignUpAPI.SignUpService))),
		),

		fx.Provide(
			healthAPI.Build,
			v1SignUpAPI.Build,
			v1SignUpAPIRouter.Build,
		),
	}
}
