package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/database"
	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/internal/svc/iam-jwks"
)

const Name = "iam-jwks"

var Module = fx.Module(Name,
	fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	}),
	database.Module,
	svc.Module,
)
