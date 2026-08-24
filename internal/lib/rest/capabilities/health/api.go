package health

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Probes holds the three probe handlers NewAPI requires. Named fields
// instead of positional arguments - Startup, Liveness, and Readiness are all
// http.HandlerFunc, and nothing stops a positional call from silently
// swapping two of them.
type Probes struct {
	Startup   http.HandlerFunc
	Liveness  http.HandlerFunc
	Readiness http.HandlerFunc
}

// NewAPI returns a ready-to-use health check API with a service discovery
// endpoint included. Health and service discovery are best friends, like
// butter and bread - both rock individually, but together, they're incredible.
//
// Example:
//
//	healthAPI, err := health.NewAPI(health.Probes{
//		Startup:   health.DefaultHandler,
//		Liveness:  health.DefaultHandler,
//		Readiness: health.DefaultHandler,
//	})
func NewAPI(probes Probes, configOptions ...Option) (func(chi.Router), error) {
	cfg := DefaultConfig()
	for _, opt := range configOptions {
		opt(cfg)
	}

	if probes.Startup == nil {
		return nil, ErrStartupRequired
	}
	if probes.Liveness == nil {
		return nil, ErrLivenessRequired
	}
	if probes.Readiness == nil {
		return nil, ErrReadinessRequired
	}
	if cfg.logger == nil {
		// NOTE: want better logs? provide a configured logger ¯\_(ツ)_/¯
		cfg.logger = slog.Default()
	}

	return func(api chi.Router) {
		api.Get("/discovery", discoveryHandler(cfg))
		api.Get("/healthz", dependencyChecksHandler(cfg))
		api.Get("/startupz", probes.Startup)
		api.Get("/livez", probes.Liveness)
		api.Get("/readyz", probes.Readiness)
	}, nil
}
