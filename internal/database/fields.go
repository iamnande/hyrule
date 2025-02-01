package database

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/iamnande/hyrule/internal/partition"
)

const (
	FieldID         = "ID"
	FieldPrimaryKey = "PK"
	FieldSortKey    = "SK"

	FieldCreatedAt = "CreatedAt"
	FieldUpdatedAt = "UpdatedAt"
	FieldDeletedAt = "DeletedAt"
	FieldExpiresAt = "ExpiresAt"
)

type TimestampFields struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (fields TimestampFields) Marshal(item map[string]types.AttributeValue) map[string]types.AttributeValue {
	item[FieldCreatedAt] = &types.AttributeValueMemberS{
		Value: fields.CreatedAt.Format(time.RFC3339),
	}
	item[FieldUpdatedAt] = &types.AttributeValueMemberS{
		Value: fields.UpdatedAt.Format(time.RFC3339),
	}
	if fields.DeletedAt != nil {
		item[FieldDeletedAt] = &types.AttributeValueMemberS{
			Value: fields.DeletedAt.Format(time.RFC3339),
		}
	}
	return item
}

func ParseTimestampFields(record map[string]types.AttributeValue) (TimestampFields, error) {
	var (
		err    error
		fields TimestampFields
	)

	fields.CreatedAt, err = ParseTimestampField(record, FieldCreatedAt)
	if err != nil {
		return TimestampFields{}, fmt.Errorf("failed to parse %s at: %w", FieldCreatedAt, err)
	}
	fields.UpdatedAt, err = ParseTimestampField(record, FieldUpdatedAt)
	if err != nil {
		return TimestampFields{}, fmt.Errorf("failed to parse %s at: %w", FieldUpdatedAt, err)
	}
	fields.DeletedAt, err = ParseNullableTimestampField(record, FieldDeletedAt)
	if err != nil {
		return TimestampFields{}, fmt.Errorf("failed to parse %s at: %w", FieldDeletedAt, err)
	}

	return fields, nil
}

func ParseTimestampField(record map[string]types.AttributeValue, field string) (time.Time, error) {
	if f, exists := record[field].(*types.AttributeValueMemberS); exists {
		ts, err := time.Parse(time.RFC3339, f.Value)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse deleted at: %w", err)
		}
		return ts, nil
	}
	return time.Time{}, fmt.Errorf("failed to parse deleted at: %v", record[field])
}

func ParseNullableTimestampField(record map[string]types.AttributeValue, field string) (*time.Time, error) {
	if f, exists := record[field].(*types.AttributeValueMemberS); exists {
		ts, err := time.Parse(time.RFC3339, f.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse deleted at: %w", err)
		}
		return &ts, nil
	}
	return nil, nil
}

func ParsePartitionField(record map[string]types.AttributeValue, field string) (partition.Partition, error) {
	var err error
	if f, exists := record[field].(*types.AttributeValueMemberS); exists {
		var sk partition.Partition
		sk, err = partition.Parse(f.Value)
		if err != nil {
			return partition.Partition{}, fmt.Errorf("failed to parse sort key: %w", err)
		}
		return sk, nil
	} else {
		return partition.Partition{}, fmt.Errorf("failed to parse sort key: %v", record[FieldSortKey])
	}
}
