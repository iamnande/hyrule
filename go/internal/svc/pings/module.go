package pings

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/svc/pings/repository"
)

var Module = fx.Module("pings",
	fx.Provide(
		repository.NewPostgres,
		newRegistry,
		newAPIHandler,
	),
)
