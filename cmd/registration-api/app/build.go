package app

import (
	"github.com/iamnande/hyrule/internal/services/password"
	"go.uber.org/fx"

	healthAPI "github.com/iamnande/hyrule/internal/apis/registration-api/health"
	v1RegistrationAPIRouter "github.com/iamnande/hyrule/internal/apis/registration-api/v1"
	v1RegistrationAPI "github.com/iamnande/hyrule/internal/apis/registration-api/v1/registration"
	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/database"
	"github.com/iamnande/hyrule/internal/domains/registration"
	registrationService "github.com/iamnande/hyrule/internal/services/registration"
)

func Build() []fx.Option {
	return []fx.Option{
		config.RegistrationAPIModule,

		// service layer
		fx.Provide(
			fx.Annotate(registrationService.NewService, fx.As(new(registration.RegistrationService))),
			fx.Annotate(password.NewService, fx.As(new(registrationService.PasswordService))),
		),

		// data layer
		fx.Provide(
			// database client
			fx.Annotate(
				database.NewDatabaseClient,
				fx.As(new(healthAPI.DatabaseClient)),
				fx.As(new(registrationService.DatabaseClient)),
			),
		),

		// domain layer
		fx.Provide(
			fx.Annotate(registration.NewDomain, fx.As(new(v1RegistrationAPI.RegistrationDomain))),
		),

		// runtime
		fx.Provide(
			healthAPI.NewHealthAPI,
			v1RegistrationAPI.NewRegistrationAPI,
			v1RegistrationAPIRouter.NewAdminAPIRouter,
		),
	}
}
