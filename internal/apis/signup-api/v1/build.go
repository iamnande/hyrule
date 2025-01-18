package v1

import (
	"log/slog"
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/rest"
	"github.com/iamnande/hyrule/internal/rest/middleware"
)

type Params struct {
	fx.In

	Logger *slog.Logger

	HealthAPI rest.APIHandler   `name:"signup-api:v1:health"`
	APIs      []rest.APIHandler `group:"signup-api:v1:apis"`
}

func Build(params Params) (http.Handler, error) {
	// TODO: default router
	router := chi.NewRouter()
	sentryHandler := sentryhttp.New(sentryhttp.Options{})
	router.Use(middleware.RequestLogger(params.Logger))
	// TODO: router.Use(REQUEST_FEATURE_FLAG)
	router.Use(middleware.ResponsePanicRecovery)
	router.Use(middleware.ResponseLogger)
	router.Mount(params.HealthAPI.URLPath(), sentryHandler.Handle(params.HealthAPI.Handler()))
	router.Route("/v1", func(v1 chi.Router) {
		for _, api := range params.APIs {
			v1.Mount(api.URLPath(), sentryHandler.Handle(api.Handler()))
		}
	})
	return router, nil
}
