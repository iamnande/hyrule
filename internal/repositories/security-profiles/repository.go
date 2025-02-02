package securityprofiles

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/getsentry/sentry-go"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/database"
	"github.com/iamnande/hyrule/internal/partition"
)

var (
	ErrNotFound = fmt.Errorf("security profile not found")

	securityProfilePartition = partition.Partition{
		Category: partition.CategoryUser,
		ID:       "security-profile",
	}
)

type DatabaseClient interface {
	database.GetItem
	database.UpdateItem
}

type Repository struct {
	db     DatabaseClient
	config config.Database
}

type Params struct {
	fx.In

	DatabaseClient DatabaseClient
	DatabaseConfig config.Database
}

func NewRepository(params Params) *Repository {
	return &Repository{
		db:     params.DatabaseClient,
		config: params.DatabaseConfig,
	}
}

type SetVerifiedRequest struct {
	Partition partition.Partition
	Verified  bool
}

// SetVerified updates the user's security profile verified status.
// Typically, this is used to mark a user's security profile as verified after
// they have verified their email address during registration. In the event
// they change their email address, this will need to be reset and re-verified.
func (repo *Repository) SetVerified(ctx context.Context, request *SetVerifiedRequest) error {
	var (
		err error

		span = sentry.StartSpan(ctx, "repositories:security-profiles:SetVerified")
	)
	defer span.Finish()
	span.SetTag("request.partition", request.Partition.String())
	span.SetTag("request.verified", fmt.Sprintf("%v", request.Verified))

	record := &Record{
		PK: request.Partition,
		SK: securityProfilePartition,
	}

	input := &dynamodb.UpdateItemInput{
		TableName: &repo.config.Name,
		Key:       record.PrimaryKey(),
		ExpressionAttributeNames: map[string]string{
			"#verified": FieldVerified,
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":verified": &types.AttributeValueMemberBOOL{
				Value: request.Verified,
			},
		},
		UpdateExpression: aws.String("SET #verified = :verified"),
	}

	if _, err = repo.db.UpdateItem(ctx, input); err != nil {
		return fmt.Errorf("failed to update security profile verified status: %w", err)
	}

	return nil
}

type DescribeSecurityProfileRequest struct {
	Partition partition.Partition
}

// DescribeSecurityProfile retrieves the security profile for the given user.
func (repo *Repository) DescribeSecurityProfile(ctx context.Context, request *DescribeSecurityProfileRequest) (*Record, error) {
	var (
		err error

		span = sentry.StartSpan(ctx, "repositories:security-profiles:DescribeSecurityProfile")
	)
	defer span.Finish()
	span.SetTag("request.partition", request.Partition.String())

	record := &Record{
		PK: request.Partition,
		SK: securityProfilePartition,
	}

	input := &dynamodb.GetItemInput{
		TableName: &repo.config.Name,
		Key:       record.PrimaryKey(),
	}

	result, err := repo.db.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get security profile: %w", err)
	}

	if result.Item == nil {
		return nil, ErrNotFound
	}

	record, err = Unmarshal(result.Item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal security profile: %w", err)
	}

	return record, nil
}
