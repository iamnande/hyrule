package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
)

const Name = "pings"

var Module = fx.Module(Name,
	fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	}),
)
