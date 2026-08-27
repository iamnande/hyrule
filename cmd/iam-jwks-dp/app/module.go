package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/internal/svc/iam-jwks-dp"
)

const Name = "iam-jwks-dp"

var Module = fx.Module(Name,
	fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	}),
	svc.Module,
)
