package health

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/rest"
	"github.com/iamnande/hyrule/internal/rest/apis/health"
	"github.com/iamnande/hyrule/internal/version"
	"go.uber.org/fx"
)

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

	API rest.APIHandler `name:"signup-api:v1:health"`
}

type Params struct {
	fx.In

	Logger     *slog.Logger
	Deployment config.Deployment

	// TODO: data layer check
	// TODO: cache layer check
}

func Build(params Params) (Result, error) {
	params.Logger.Info("building health API")
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
				return nil
			},
		),
		health.WithSoftDependency("cache",
			func(ctx context.Context) error {
				return nil
			},
		),
	)
	if err != nil {
		return Result{}, err
	}
	params.Logger.Info("built health API")
	return Result{
		API: &API{
			handler: handler,
		},
	}, nil
}
