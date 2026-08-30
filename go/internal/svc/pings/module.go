package pings

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest"
	pingsapi "github.com/iamnande/hyrule/go/internal/svc/pings/api"
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
	"github.com/iamnande/hyrule/go/internal/svc/pings/repository"
)

var Module = fx.Module("pings",
	fx.Provide(
		repository.NewPostgres,
		newRegistry,
		newAPIHandler,
	),
)

type apiHandlerResult struct {
	fx.Out

	API rest.APIHandler `group:"apis"`
}

func newRegistry(repo *repository.PostgresRepository, cfg domain.Config) *domain.Registry {
	return domain.NewRegistry(repo, cfg)
}

func newAPIHandler(service *domain.Registry) apiHandlerResult {
	return apiHandlerResult{
		API: pingsapi.NewRouter(pingsapi.NewHandlers(service)),
	}
}
