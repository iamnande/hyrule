package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest/capabilities/health"
	svc "github.com/iamnande/hyrule/go/internal/svc/iam-jwks"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/repository"
)

const Name = "iam-jwks"

var Module = fx.Module(Name,
	fx.Supply(health.DefaultProbes),
	fx.Provide(repository.LoadEnvConfig),
	svc.Module,
)
