package userbillingprofile

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

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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
	val[database.FieldCreatedAt] = &types.AttributeValueMemberS{
		Value: record.CreatedAt.Format(time.RFC3339),
	}
	val[database.FieldUpdatedAt] = &types.AttributeValueMemberS{
		Value: record.UpdatedAt.Format(time.RFC3339),
	}
	if record.DeletedAt != nil {
		val[database.FieldDeletedAt] = &types.AttributeValueMemberS{
			Value: record.DeletedAt.Format(time.RFC3339),
		}
	}
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
	record.CreatedAt, err = database.ParseTimestampField(item, database.FieldCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created at: %w", err)
	}
	record.UpdatedAt, err = database.ParseTimestampField(item, database.FieldUpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated at: %w", err)
	}
	record.DeletedAt, err = database.ParseNullableTimestampField(item, database.FieldDeletedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deleted at: %w", err)
	}

	return record, nil
}
