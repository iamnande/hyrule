package v1

import (
	"log/slog"
	"net/http"

	"github.com/iamnande/hyrule/internal/rest/router"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/rest"
)

type Params struct {
	fx.In

	Logger *slog.Logger

	HealthAPI rest.APIHandler   `name:"signup-api:v1:health"`
	APIs      []rest.APIHandler `group:"signup-api:v1:apis"`
}

func Build(params Params) (http.Handler, error) {
	return router.NewRouter(&router.Config{
		Logger:      params.Logger,
		HealthAPI:   params.HealthAPI,
		APIs:        params.APIs,
		VersionPath: "/v1",
	}), nil
}
