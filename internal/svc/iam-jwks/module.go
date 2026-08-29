package iamjwks

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/rest"
	jwksapi "github.com/iamnande/hyrule/internal/svc/iam-jwks/api"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/domain"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/repository"
)

func WithPostgres() fx.Option {
	return fx.Provide(repository.New, newServiceFromPostgres, newAPIHandler)
}

func WithFileStore() fx.Option {
	return fx.Provide(repository.NewFile, newServiceFromFile, newAPIHandler)
}

type apiHandlerResult struct {
	fx.Out

	API rest.APIHandler `group:"apis"`
}

func newServiceFromPostgres(repo *repository.Repository) *domain.Service {
	return domain.NewService(repo)
}

func newServiceFromFile(repo *repository.FileRepository) *domain.Service {
	return domain.NewService(repo)
}

func newAPIHandler(service *domain.Service) apiHandlerResult {
	return apiHandlerResult{
		API: jwksapi.NewRouter(jwksapi.NewHandlers(service)),
	}
}
