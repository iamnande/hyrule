package registration

import (
	"context"
	"log/slog"

	"github.com/getsentry/sentry-go"
	"github.com/iamnande/hyrule/internal/models"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/services/registration"
)

type RegistrationService interface {
	RegisterNewUser(
		ctx context.Context,
		input *registration.RegisterNewUserInput,
	) (*registration.RegisterNewUserOutput, error)
}

type Domain struct {
	logger              *slog.Logger
	registrationService RegistrationService
}

type Params struct {
	fx.In

	Logger              *slog.Logger
	RegistrationService RegistrationService
}

func NewDomain(params Params) *Domain {
	return &Domain{
		logger:              params.Logger,
		registrationService: params.RegistrationService,
	}
}

type RegisterNewUserInput struct {
	Email    string
	Password string
	FullName string
}

type RegisterNewUserOutput struct {
	User models.User
	// SecurityProfile models.SecurityProfile
	// BillingProfile models.BillingProfile
}

func (domain *Domain) RegisterNewUser(
	ctx context.Context,
	input *RegisterNewUserInput,
) (*RegisterNewUserOutput, error) {
	span := sentry.StartSpan(ctx, "domains:registration:RegisterNewUser")
	defer span.Finish()
	result, err := domain.registrationService.RegisterNewUser(span.Context(), &registration.RegisterNewUserInput{
		Email:    input.Email,
		FullName: input.FullName,
		Password: input.Password,
	})
	if err != nil {
		return nil, err
	}
	return &RegisterNewUserOutput{
		User: result.User,
	}, nil
}
