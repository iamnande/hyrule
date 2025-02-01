package registration

import (
	"context"

	"github.com/getsentry/sentry-go"

	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/services/notification"
	"github.com/iamnande/hyrule/internal/services/registration"
)

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
	var (
		err error

		span = sentry.StartSpan(ctx, "domains:registration:RegisterNewUser")
	)
	defer span.Finish()

	var userRegistration *registration.RegisterNewUserOutput
	userRegistration, err = domain.registrationService.RegisterNewUser(span.Context(), &registration.RegisterNewUserInput{
		Email:    input.Email,
		FullName: input.FullName,
		Password: input.Password,
	})
	if err != nil {
		return nil, err
	}

	// Email notification
	// type=email,templateName=ACCOUNT_VERIFICATION
	if err = domain.notificationService.NotifyEmail(span.Context(), &notification.NotifyEmailInput{
		Recipient: userRegistration.User.Email,

		Template: notification.TemplateAccountVerification,
		Metadata: map[string]string{
			"user.id":    userRegistration.User.ID.String(),
			"user.email": userRegistration.User.Email,
			"user.name":  userRegistration.User.FullName,
			"plan.name":  userRegistration.Plan.String(),
			"verify.url": "http://localhost:8000/v1/verify?token=" + "token",
		},
	}); err != nil {
		return nil, err
	}

	return &RegisterNewUserOutput{
		User: userRegistration.User,
	}, nil
}
