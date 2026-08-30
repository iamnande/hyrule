package app

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/config"
	"github.com/iamnande/hyrule/go/internal/lib/database"
	"github.com/iamnande/hyrule/go/internal/lib/rest/capabilities/health"
	"github.com/iamnande/hyrule/go/internal/svc/pings"
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
)

const Name = "pings"

var Module = fx.Module(Name,
	fx.Supply(health.DefaultProbes),
	fx.Provide(config.LoadDatabase()),
	fx.Provide(config.Load[domain.Config]("PINGS")),
	database.Module,
	pings.Module,
)
