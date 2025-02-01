package billingprofile

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/iamnande/hyrule/internal/models"

	"github.com/iamnande/hyrule/internal/database"
	"github.com/iamnande/hyrule/internal/partition"
)

const (
	FieldPlan = "UserBillingPlan"
)

type Record struct {
	PK partition.Partition
	SK partition.Partition

	Plan models.BillingPlan

	database.TimestampFields
}

type NewBillingProfileParams struct {
	Partition partition.Partition
	Plan      models.BillingPlan
}

func NewBillingProfile(params NewBillingProfileParams) *Record {
	now := time.Now().UTC()
	return &Record{
		PK: params.Partition,
		SK: partition.Partition{
			Category: partition.CategoryAccount,
			ID:       "billing-profile",
		},
		Plan: models.BillingPlanConsumption,
		TimestampFields: database.TimestampFields{
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func (record *Record) PrimaryKey() map[string]types.AttributeValue {
	key := map[string]types.AttributeValue{
		database.PartitionKeyAttributeName: &types.AttributeValueMemberS{
			Value: record.PK.String(),
		},
		database.SortKeyAttributeName: &types.AttributeValueMemberS{
			Value: record.SK.String(),
		},
	}
	return key
}

func (record *Record) Marshal() map[string]types.AttributeValue {
	val := record.PrimaryKey()
	val[FieldPlan] = &types.AttributeValueMemberS{
		Value: string(record.Plan),
	}
	record.TimestampFields.Marshal(val)
	return val
}

func Unmarshal(item map[string]types.AttributeValue) (*Record, error) {
	var (
		err    error
		record = &Record{}
	)

	// PK
	record.PK, err = database.ParsePartitionField(item, database.FieldPrimaryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse partition key: %w", err)
	}

	// SK
	record.SK, err = database.ParsePartitionField(item, database.FieldSortKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sort key: %w", err)
	}

	// Plan
	if field, exists := item[FieldPlan].(*types.AttributeValueMemberS); exists {
		record.Plan = models.BillingPlan(field.Value)
	} else {
		return nil, fmt.Errorf("failed to parse plan: %v", item[FieldPlan])
	}

	// Timestamps
	timestampFields, err := database.ParseTimestampFields(item)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamps: %w", err)
	}
	record.TimestampFields = timestampFields

	return record, nil
}
