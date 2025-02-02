package health

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/rest"
	"github.com/iamnande/hyrule/internal/rest/apis/health"
	"github.com/iamnande/hyrule/internal/version"
)

type DatabaseClient interface {
	dynamodb.ScanAPIClient
}

type API struct {
	handler http.Handler
}

func (api *API) URLPath() string {
	return "/"
}

func (api *API) Handler() http.Handler {
	return api.handler
}

type Result struct {
	fx.Out

	API rest.APIHandler `name:"invite-api:v1:health"`
}

type Params struct {
	fx.In

	Logger     *slog.Logger
	Deployment config.Deployment

	DatabaseClient DatabaseClient
	DatabaseConfig config.Database
}

func NewHealthAPI(params Params) (Result, error) {
	handler, err := health.NewAPI(
		health.DefaultHandler, // liveness
		health.DefaultHandler, // readiness
		health.WithLogger(params.Logger),
		health.WithServiceMetadata(&health.ServiceMetadata{
			Name:        version.ServiceName,
			Version:     version.ServiceVersion,
			Commit:      version.ServiceCommit,
			Environment: string(params.Deployment.Environment),
			Region:      string(params.Deployment.Region),
		}),
		health.WithHardDependency("database",
			func(ctx context.Context) error {
				if _, err := params.DatabaseClient.Scan(ctx, &dynamodb.ScanInput{
					TableName: aws.String(params.DatabaseConfig.Name),
					Limit:     aws.Int32(1),
				}); err != nil {
					params.Logger.Error("failed to check database health", slog.Any("error", err))
					return err
				}
				return nil
			},
		),
	)
	if err != nil {
		return Result{}, err
	}
	return Result{
		API: &API{
			handler: handler,
		},
	}, nil
}
