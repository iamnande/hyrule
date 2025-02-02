package app

import (
	"go.uber.org/fx"

	healthAPI "github.com/iamnande/hyrule/internal/apis/registration-api/health"
	v1RegistrationAPIRouter "github.com/iamnande/hyrule/internal/apis/registration-api/v1"
	v1RegistrationAPI "github.com/iamnande/hyrule/internal/apis/registration-api/v1/registration"
	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/database"
	registrationDomain "github.com/iamnande/hyrule/internal/domains/registration"
	"github.com/iamnande/hyrule/internal/services/notification"
	"github.com/iamnande/hyrule/internal/services/password"
	"github.com/iamnande/hyrule/internal/services/registration"
)

func Build() []fx.Option {
	return []fx.Option{
		config.RegistrationAPIModule,

		// service layer
		fx.Provide(
			fx.Annotate(password.NewService,
				fx.As(new(registration.PasswordService)),
			),
			fx.Annotate(registration.NewService,
				fx.As(new(registrationDomain.RegistrationService)),
			),
			fx.Annotate(notification.NewService,
				fx.As(new(registrationDomain.NotificationService)),
			),
		),

		// data layer
		fx.Provide(
			// database client
			fx.Annotate(database.NewDatabaseClient,
				fx.As(new(healthAPI.DatabaseClient)),
				fx.As(new(registration.DatabaseClient)),
			),
		),

		// domain layer
		fx.Provide(
			fx.Annotate(registrationDomain.NewDomain,
				fx.As(new(v1RegistrationAPI.RegistrationDomain)),
			),
		),

		// runtime
		fx.Provide(
			healthAPI.NewHealthAPI,
			v1RegistrationAPI.NewRegistrationAPI,
			v1RegistrationAPIRouter.NewRegistrationAPIRouter,
		),
	}
}
