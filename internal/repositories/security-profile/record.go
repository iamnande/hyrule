package securityprofile

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

	database.TimestampFields
}

type NewSecurityProfileParams struct {
	Partition partition.Partition
	Password  string
}

func NewSecurityProfile(params NewSecurityProfileParams) *Record {
	now := time.Now().UTC()
	return &Record{
		PK: params.Partition,
		SK: partition.Partition{
			Category: partition.CategoryUser,
			ID:       "security-profile",
		},
		Password: params.Password,
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
	val[FieldPassword] = &types.AttributeValueMemberS{
		Value: record.Password,
	}
	val[FieldVerified] = &types.AttributeValueMemberBOOL{
		Value: record.Verified,
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
	timestampFields, err := database.ParseTimestampFields(item)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamps: %w", err)
	}
	record.TimestampFields = timestampFields

	return record, nil
}
