package user

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/segmentio/ksuid"

	"github.com/iamnande/hyrule/internal/database"
	"github.com/iamnande/hyrule/internal/partition"
)

const (
	FieldID       = "UserID"
	FieldEmail    = "UserEmail"
	FieldFullName = "UserFullName"
)

type Record struct {
	PK partition.Partition
	SK partition.Partition

	ID       ksuid.KSUID
	Email    string
	FullName string

	database.TimestampFields
}

type NewUserParams struct {
	Email    string
	FullName string
}

func NewUser(params NewUserParams) *Record {
	id := ksuid.New()
	now := time.Now().UTC()
	return &Record{
		PK: partition.Partition{
			Category: partition.CategoryUser,
			ID:       id.String(),
		},
		SK: partition.Partition{
			Category: partition.CategoryUser,
			ID:       "profile",
		},
		ID:       id,
		Email:    params.Email,
		FullName: params.FullName,
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
	val[FieldID] = &types.AttributeValueMemberS{
		Value: record.ID.String(),
	}
	val[FieldEmail] = &types.AttributeValueMemberS{
		Value: record.Email,
	}
	val[FieldFullName] = &types.AttributeValueMemberS{
		Value: record.FullName,
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
	record.Email = record.PK.ID

	// SK
	record.SK, err = database.ParsePartitionField(item, database.FieldSortKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sort key: %w", err)
	}

	// ID
	if field, exists := item[database.FieldID].(*types.AttributeValueMemberS); exists {
		var id ksuid.KSUID
		id, err = ksuid.Parse(field.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse id: %w", err)
		}
		record.ID = id
	} else {
		return nil, fmt.Errorf("failed to parse id: %v", item[database.FieldID])
	}

	// Email
	if field, exists := item[FieldEmail].(*types.AttributeValueMemberS); exists {
		record.Email = field.Value
	} else {
		return nil, fmt.Errorf("failed to parse email: %v", item[FieldEmail])
	}

	// FullName
	if field, exists := item[FieldFullName].(*types.AttributeValueMemberS); exists {
		record.FullName = field.Value
	} else {
		return nil, fmt.Errorf("failed to parse full name: %v", item[FieldFullName])
	}

	// Timestamps
	timestampFields, err := database.ParseTimestampFields(item)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamps: %w", err)
	}
	record.TimestampFields = timestampFields

	return record, nil
}
