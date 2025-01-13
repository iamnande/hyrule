package health

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewAPI returns a ready-to-use health check API with a service discovery
// endpoint included. Health and service discovery are best friends, like
// butter and bread - both rock individually, but together, they're incredible.
//
// Example:
//
//	healthAPI, err := health.NewAPI(health.DefaultHandler, health.DefaultHandler)
func NewAPI(liveness, readiness http.HandlerFunc, configOptions ...Option) (http.Handler, error) {
	cfg := DefaultConfig()
	for _, opt := range configOptions {
		opt(cfg)
	}

	if liveness == nil {
		return nil, ErrLivenessRequired
	}
	if readiness == nil {
		return nil, ErrReadinessRequired
	}
	if cfg.logger == nil {
		// NOTE: want better logs? provide a logger ¯\_(ツ)_/¯
		cfg.logger = slog.Default()
	}

	api := chi.NewRouter()
	// URL Path: /discovery
	api.Get("/discovery", discoveryHandler(cfg))
	// URL Path: /health
	api.Route("/health", func(health chi.Router) {
		// URL Path: /health/dependencies
		health.Get("/dependencies", dependencyChecksHandler(cfg))
		// URL Path: /health/probes
		health.Route("/probes", func(probes chi.Router) {
			// URL Path: /health/probes/liveness
			probes.Get("/liveness", liveness)
			// URL Path: /health/probes/readiness
			probes.Get("/readiness", readiness)
		})
	})

	return api, nil
}
