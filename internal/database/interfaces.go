package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type GetItem interface {
	GetItem(
		ctx context.Context,
		params *dynamodb.GetItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.GetItemOutput, error)
}

type PutItem interface {
	PutItem(
		ctx context.Context,
		params *dynamodb.PutItemInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.PutItemOutput, error)
}

type TransactWriteItems interface {
	TransactWriteItems(
		ctx context.Context,
		params *dynamodb.TransactWriteItemsInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.TransactWriteItemsOutput, error)
}
