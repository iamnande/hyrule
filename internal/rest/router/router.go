package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/iamnande/hyrule/internal/rest"
	"github.com/iamnande/hyrule/internal/rest/middleware"
)

type Config struct {
	Logger *slog.Logger

	HealthAPI   rest.APIHandler
	APIs        []rest.APIHandler
	VersionPath string // e.g. /v1, /v2, etc.
}

func NewRouter(cfg *Config) http.Handler {
	// TODO: default router
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(cfg.Logger))
	router.Use(middleware.RequestTracer)
	// TODO: router.Use(REQUEST_FEATURE_FLAG)
	router.Use(middleware.ResponsePanicRecovery)
	router.Use(middleware.ResponseLogger)
	router.Mount(cfg.HealthAPI.URLPath(), cfg.HealthAPI.Handler())
	router.Route(cfg.VersionPath, func(path chi.Router) {
		for _, api := range cfg.APIs {
			path.Mount(api.URLPath(), api.Handler())
		}
	})
	return router
}
