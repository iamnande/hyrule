package registration

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/getsentry/sentry-go"
	"github.com/segmentio/ksuid"

	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/repositories/billing-profiles"
	"github.com/iamnande/hyrule/internal/repositories/invites"
	"github.com/iamnande/hyrule/internal/repositories/security-profiles"
	"github.com/iamnande/hyrule/internal/repositories/users"
)

type RegisterNewUserInput struct {
	FullName string
	Email    string
	Password string
}

type RegisterNewUserOutput struct {
	User        models.User
	Plan        models.BillingPlan
	InviteToken ksuid.KSUID
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
	userRecord := users.NewUser(users.NewUserParams{
		Email:    input.Email,
		FullName: input.FullName,
	})

	// security profile
	securityProfileRecord := securityprofiles.NewSecurityProfile(securityprofiles.NewSecurityProfileParams{
		Partition: userRecord.Partition(),
		Password:  hashedPassword,
	})

	// billing profile
	billingProfileRecord := billingprofiles.NewBillingProfile(billingprofiles.NewBillingProfileParams{
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

	return &RegisterNewUserOutput{
		User:        models.MarshalUser(userRecord),
		Plan:        billingProfileRecord.Plan,
		InviteToken: inviteRecord.Token,
	}, nil
}
