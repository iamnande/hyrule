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
)

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
