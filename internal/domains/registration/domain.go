package registration

import (
	"context"
	"log/slog"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/services/notification"
	"github.com/iamnande/hyrule/internal/services/registration"
)

type RegistrationService interface {
	RegisterNewUser(
		ctx context.Context,
		input *registration.RegisterNewUserInput,
	) (*registration.RegisterNewUserOutput, error)
}

type NotificationService interface {
	NotifyEmail(ctx context.Context, input *notification.NotifyEmailInput) error
}

type Domain struct {
	logger              *slog.Logger
	registrationService RegistrationService
	notificationService NotificationService
}

type Params struct {
	fx.In

	Logger              *slog.Logger
	RegistrationService RegistrationService
	NotificationService NotificationService
}

func NewDomain(params Params) *Domain {
	return &Domain{
		logger:              params.Logger,
		registrationService: params.RegistrationService,
		notificationService: params.NotificationService,
	}
}
