package pings

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/rest"
	pingsapi "github.com/iamnande/hyrule/internal/svc/pings/api"
	"github.com/iamnande/hyrule/internal/svc/pings/domain"
	"github.com/iamnande/hyrule/internal/svc/pings/repository"
)

var Module = fx.Module("pings",
	fx.Provide(
		repository.New,
		newService,
		newAPIHandler,
	),
)

type apiHandlerResult struct {
	fx.Out

	API rest.APIHandler `group:"apis"`
}

func newService(repo *repository.Repository, cfg domain.Config) *domain.Service {
	return domain.NewService(repo, cfg)
}

func newAPIHandler(service *domain.Service) apiHandlerResult {
	return apiHandlerResult{
		API: pingsapi.NewRouter(pingsapi.NewHandlers(service)),
	}
}
