package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
)

type Params struct {
	fx.In

	Config     config.Database
	Deployment config.Deployment
}

func NewDatabaseClient(params Params) (*dynamodb.Client, error) {
	awsConfig, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion(params.Deployment.Region.String()),
	)
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(awsConfig, func(opts *dynamodb.Options) {
		if params.Deployment.Environment == config.LocalEnvironment {
			opts.BaseEndpoint = aws.String(params.Config.LocalEndpoint)
			opts.Region = params.Deployment.Region.String()
			opts.Credentials = credentials.NewStaticCredentialsProvider("local", "local", "local")
		}
	}), nil
}
