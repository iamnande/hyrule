package invites

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/segmentio/ksuid"

	"github.com/iamnande/hyrule/internal/database"
	"github.com/iamnande/hyrule/internal/partition"
)

const (
	FieldStatus = "InviteStatus"
	FieldToken  = "InviteToken"
)

type Status string

const (
	StatusPending  Status = "PENDING"
	StatusAccepted Status = "ACCEPTED"
	StatusExpired  Status = "EXPIRED"
)

type Record struct {
	PK partition.Partition
	SK partition.Partition

	Status    Status
	ExpiresAt time.Time
	Token     ksuid.KSUID

	database.TimestampFields
}

type NewInviteParams struct {
	Partition partition.Partition
	ExpiresAt time.Time
}

func NewInvite(params NewInviteParams) *Record {
	token := ksuid.New()
	now := time.Now().UTC()
	return &Record{
		PK: params.Partition,
		SK: partition.Partition{
			Category: partition.CategoryInvite,
			ID:       token.String(),
		},

		Token:     token,
		Status:    StatusPending,
		ExpiresAt: params.ExpiresAt,

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
	val[FieldStatus] = &types.AttributeValueMemberS{
		Value: string(record.Status),
	}
	val[FieldToken] = &types.AttributeValueMemberS{
		Value: record.Token.String(),
	}
	record.TimestampFields.Marshal(val)
	return val
}
