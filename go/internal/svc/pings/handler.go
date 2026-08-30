package pings

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest"
	pingsapi "github.com/iamnande/hyrule/go/internal/svc/pings/api"
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
)

type apiHandlerResult struct {
	fx.Out

	API rest.APIHandler `group:"apis"`
}

func newAPIHandler(service *domain.Registry) apiHandlerResult {
	return apiHandlerResult{
		API: pingsapi.NewRouter(pingsapi.NewHandlers(service)),
	}
}
