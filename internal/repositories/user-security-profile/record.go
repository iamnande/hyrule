package usersecurityprofile

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/iamnande/hyrule/internal/database"
	"github.com/iamnande/hyrule/internal/partition"
)

const (
	FieldPassword = "UserSecurityPassword"
	FieldVerified = "UserSecurityVerified"
)

type Record struct {
	PK partition.Partition
	SK partition.Partition

	Password string
	Verified bool

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
	val[FieldPassword] = &types.AttributeValueMemberS{
		Value: record.Password,
	}
	val[FieldVerified] = &types.AttributeValueMemberBOOL{
		Value: record.Verified,
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

	// Password
	if field, exists := item[FieldPassword].(*types.AttributeValueMemberS); exists {
		record.Password = field.Value
	} else {
		return nil, fmt.Errorf("failed to parse password: %v", item[FieldPassword])
	}

	// Verified
	if field, exists := item[FieldVerified].(*types.AttributeValueMemberBOOL); exists {
		record.Verified = field.Value
	} else {
		return nil, fmt.Errorf("failed to parse verified: %v", item[FieldVerified])
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
