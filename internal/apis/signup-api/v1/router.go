package v1

import (
	"log/slog"
	"net/http"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/rest"
	"github.com/iamnande/hyrule/internal/rest/router"
)

type Params struct {
	fx.In

	Logger *slog.Logger

	HealthAPI rest.APIHandler   `name:"signup-api:v1:health"`
	APIs      []rest.APIHandler `group:"signup-api:v1:apis"`
}

func NewSignUpAPIRouter(params Params) (http.Handler, error) {
	return router.NewRouter(&router.Config{
		Logger:      params.Logger,
		HealthAPI:   params.HealthAPI,
		APIs:        params.APIs,
		VersionPath: "/v1",
	}), nil
}
