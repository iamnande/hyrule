package v1

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/rest"
)

type Params struct {
	fx.In

	Logger *slog.Logger

	HealthAPI rest.APIHandler   `name:"admin-api:v1:health"`
	APIs      []rest.APIHandler `group:"admin-api:v1:apis"`
}

func Build(params Params) (http.Handler, error) {
	router := chi.NewRouter()
	// TODO: router.Use(REQUEST_LOGGER)
	// TODO: router.Use(REQUEST_FEATURE_FLAG)
	// TODO: router.Use(RESPONSE_PANIC_RECOVERY)
	router.Mount(params.HealthAPI.URLPath(), params.HealthAPI.Handler())
	router.Route("/v1", func(v1 chi.Router) {
		for _, api := range params.APIs {
			v1.Mount(api.URLPath(), api.Handler())
		}
	})
	return router, nil
}
