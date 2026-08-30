package health

import (
	"net/http"
	"time"
)

var (
	DefaultTimeout = time.Second * 5

	DefaultConfig = func() *config {
		return &config{
			timeout:         DefaultTimeout,
			serviceMetadata: &ServiceMetadata{},
		}
	}

	// DefaultHandler is a no-op handler. It makes for a fair starting place
	// for kubernetes liveness and readiness probes.
	DefaultHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// DefaultProbes wires DefaultHandler into all three probes - the
	// starting point for a service with no real dependency checks yet.
	DefaultProbes = Probes{
		Startup:   DefaultHandler,
		Liveness:  DefaultHandler,
		Readiness: DefaultHandler,
	}
)
