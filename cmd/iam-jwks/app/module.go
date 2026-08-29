package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/database"
	"github.com/iamnande/hyrule/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/internal/svc/iam-jwks"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/repository"
)

const Name = "iam-jwks"

func Module(fileCfg repository.FileConfig) fx.Option {
	probes := fx.Supply(health.Probes{
		Startup:   health.DefaultHandler,
		Liveness:  health.DefaultHandler,
		Readiness: health.DefaultHandler,
	})

	if fileCfg.Path != "" {
		return fx.Module(Name, probes, fx.Supply(fileCfg), svc.WithFileStore())
	}
	return fx.Module(Name, probes, database.Module, svc.WithPostgres())
}
