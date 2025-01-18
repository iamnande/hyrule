package app

import (
	"github.com/iamnande/hyrule/internal/services/signup"
	"go.uber.org/fx"

	healthAPI "github.com/iamnande/hyrule/internal/apis/signup-api/health"
	v1SignUpAPIRouter "github.com/iamnande/hyrule/internal/apis/signup-api/v1"
	v1SignUpAPI "github.com/iamnande/hyrule/internal/apis/signup-api/v1/signup"
	"github.com/iamnande/hyrule/internal/config"
)

func Build() []fx.Option {
	return []fx.Option{
		config.SignUpAPIConfigModule,

		// TODO: tracing

		// TODO: service layer
		fx.Provide(
			fx.Annotate(signup.NewService, fx.As(new(v1SignUpAPI.SignUpService))),
		),

		// TODO: data layer
		// fx.Provide(
		// 	fx.Annotate(serviceRepository.NewRepository, fx.As(new(adminDomain.ServiceRepository))),
		// ),
		//
		// // TODO: domain layer
		// fx.Provide(
		// 	fx.Annotate(adminDomain.NewDomain, fx.As(new(v1AdminOrganizationsAPI.AdminDomain))),
		// ),

		fx.Provide(
			healthAPI.Build,
			v1SignUpAPI.Build,
			v1SignUpAPIRouter.Build,
		),
	}
}
