package registration

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/getsentry/sentry-go"
	"github.com/jordan-wright/email"

	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/repositories/billing-profile"
	"github.com/iamnande/hyrule/internal/repositories/invites"
	"github.com/iamnande/hyrule/internal/repositories/security-profile"
	"github.com/iamnande/hyrule/internal/repositories/user"
)

type RegisterNewUserInput struct {
	FullName string
	Email    string
	Password string
}

type RegisterNewUserOutput struct {
	User models.User
	Plan models.BillingPlan
}

func (service *Service) RegisterNewUser(
	ctx context.Context,
	input *RegisterNewUserInput,
) (*RegisterNewUserOutput, error) {
	span := sentry.StartSpan(ctx, "services:registration:RegisterNewUser")
	defer span.Finish()

	// before we get started, make sure we can hash the password
	hashedPassword, err := service.passwordService.HashPassword(span.Context(), input.Password)
	if err != nil {
		return nil, err
	}

	// user
	userRecord := user.NewUser(user.NewUserParams{
		Email:    input.Email,
		FullName: input.FullName,
	})

	// security profile
	securityProfileRecord := securityprofile.NewSecurityProfile(securityprofile.NewSecurityProfileParams{
		Partition: userRecord.Partition(),
		Password:  hashedPassword,
	})

	// billing profile
	billingProfileRecord := billingprofile.NewBillingProfile(billingprofile.NewBillingProfileParams{
		Partition: userRecord.Partition(),
		Plan:      models.BillingPlanConsumption,
	})

	// invite
	inviteRecord := invites.NewInvite(invites.NewInviteParams{
		Partition: userRecord.Partition(),
		ExpiresAt: time.Now().UTC().Add(service.inviteConfig.ValidityWindow),
	})

	// write all records in a single transaction
	_, err = service.databaseClient.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			// user
			{
				Put: &types.Put{
					TableName: &service.databaseConfig.Name,
					Item:      userRecord.Marshal(),
				},
			},
			// security profile
			{
				Put: &types.Put{
					TableName: &service.databaseConfig.Name,
					Item:      securityProfileRecord.Marshal(),
				},
			},
			// billing profile
			{
				Put: &types.Put{
					TableName: &service.databaseConfig.Name,
					Item:      billingProfileRecord.Marshal(),
				},
			},
			// invite
			{
				Put: &types.Put{
					TableName: &service.databaseConfig.Name,
					Item:      inviteRecord.Marshal(),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// send verification email
	// TODO: replace with an event produced to the notification outbox
	// TODO: move to registration domain via notification service
	// TODO: event outbox
	emailHost := "localhost"
	emailPort := 1025
	emailEndpoint := fmt.Sprintf("%s:%d", emailHost, emailPort)
	verificationEmail := email.NewEmail()
	verificationEmail.To = []string{input.Email}
	verificationEmail.From = "Hyrule <noreply@hyrule.com>"
	verificationEmail.Subject = "[MHQ] Welcome to Hyrule!"
	verificationEmail.HTML = []byte(`<h1>Welcome to Hyrule!</h1>
	    <p>Thank you for signing up for Hyrule. We're excited to have you on board!</p>
        <p>Please verify your account by clicking here: <a href="%s%s">Verify Account</a></p>
		<p>Thanks,</p>
		<p>The Hyrule Team</p>
	`)

	if err = verificationEmail.Send(emailEndpoint, nil); err != nil {
		return nil, err
	}

	return &RegisterNewUserOutput{
		User: models.MarshalUser(userRecord),
		Plan: billingProfileRecord.Plan,
	}, nil
}
