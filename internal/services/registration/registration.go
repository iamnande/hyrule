package registration

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/getsentry/sentry-go"
	userbillingprofile "github.com/iamnande/hyrule/internal/repositories/user-billing-profile"
	"github.com/segmentio/ksuid"

	"github.com/iamnande/hyrule/internal/models"
	"github.com/iamnande/hyrule/internal/partition"
	"github.com/iamnande/hyrule/internal/repositories/user"
	usersecrurityprofile "github.com/iamnande/hyrule/internal/repositories/user-security-profile"
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

	// prep
	userID := ksuid.New()
	createdAt := time.Now().UTC()
	userPartition := partition.Partition{
		Category: partition.CategoryUser,
		ID:       input.Email,
	}

	// user
	userRecord := &user.Record{
		PK: userPartition,
		SK: partition.Partition{
			Category: partition.CategoryUser,
			ID:       "profile",
		},
		ID:        userID,
		Email:     input.Email,
		FullName:  input.FullName,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	// security profile
	userSecurityProfileRecord := &usersecrurityprofile.Record{
		PK: userPartition,
		SK: partition.Partition{
			Category: partition.CategoryUser,
			ID:       "security-profile",
		},
		Password:  hashedPassword,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	// billing profile
	userBillingProfileRecord := &userbillingprofile.Record{
		PK: userPartition,
		SK: partition.Partition{
			Category: partition.CategoryUser,
			ID:       "billing-profile",
		},
		Plan:      models.BillingPlanConsumption,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

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
					Item:      userSecurityProfileRecord.Marshal(),
				},
			},
			// billing profile
			{
				Put: &types.Put{
					TableName: &service.databaseConfig.Name,
					Item:      userBillingProfileRecord.Marshal(),
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return &RegisterNewUserOutput{
		User: models.MarshalUser(userRecord),
		Plan: userBillingProfileRecord.Plan,
	}, nil
}
