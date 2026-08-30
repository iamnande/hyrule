package rest

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/iamnande/hyrule/go/internal/lib/rest/middleware"
)

type Config struct {
	Logger *slog.Logger

	HealthAPI APIHandler
	APIs      []APIHandler
}

// NewRouter mounts the health API and every service API flat - no version
// segment in the path. version lives in a header, not the URL (see
// docs/conventions.md#url-structure--versioning).
func NewRouter(cfg *Config) http.Handler {
	// TODO: default router
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger(cfg.Logger))
	router.Use(middleware.RequestTracer)
	// TODO: router.Use(REQUEST_FEATURE_FLAG)
	router.Use(middleware.ResponsePanicRecovery)
	router.Use(middleware.ResponseLogger)
	cfg.HealthAPI(router)
	for _, api := range cfg.APIs {
		api(router)
	}
	return router
}
