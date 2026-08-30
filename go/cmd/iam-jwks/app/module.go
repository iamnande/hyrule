package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/go/internal/svc/iam-jwks"
)

const Name = "iam-jwks"

var Module = fx.Module(Name,
	fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	}),
	svc.Module,
)
