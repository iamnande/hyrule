package app

import (
	serviceRepository "github.com/iamnande/hyrule/internal/repositories/service"
	"go.uber.org/fx"

	healthAPI "github.com/iamnande/hyrule/internal/apis/admin-api/health"
	v1AdminAPIRouter "github.com/iamnande/hyrule/internal/apis/admin-api/v1"
	v1ServiceAPI "github.com/iamnande/hyrule/internal/apis/admin-api/v1/services"
	"github.com/iamnande/hyrule/internal/config"
	adminDomain "github.com/iamnande/hyrule/internal/domains/admin"
)

func Build() []fx.Option {
	return []fx.Option{
		config.AdminAPIConfigModule,

		// TODO: tracing

		// TODO: service layer

		// TODO: data layer
		fx.Provide(
			fx.Annotate(serviceRepository.NewRepository, fx.As(new(adminDomain.ServiceRepository))),
		),

		// TODO: domain layer
		fx.Provide(
			fx.Annotate(adminDomain.NewDomain, fx.As(new(v1ServiceAPI.AdminDomain))),
		),

		fx.Provide(
			healthAPI.Build,
			v1ServiceAPI.Build,
			v1AdminAPIRouter.Build,
		),
	}
}
