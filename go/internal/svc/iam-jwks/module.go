package iamjwks

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest"
	jwksapi "github.com/iamnande/hyrule/go/internal/svc/iam-jwks/api"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/domain"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/repository"
)

var Module = fx.Module("iam-jwks",
	fx.Provide(
		repository.NewEnv,
		newService,
		newAPIHandler,
	),
)

type apiHandlerResult struct {
	fx.Out

	API rest.APIHandler `group:"apis"`
}

func newService(repo *repository.EnvRepository) *domain.Service {
	return domain.NewService(repo)
}

func newAPIHandler(service *domain.Service) apiHandlerResult {
	return apiHandlerResult{
		API: jwksapi.NewRouter(jwksapi.NewHandlers(service)),
	}
}
