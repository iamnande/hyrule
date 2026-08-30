package iamjwks

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest"
	jwksapi "github.com/iamnande/hyrule/go/internal/svc/iam-jwks/api"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/domain"
)

type apiHandlerResult struct {
	fx.Out

	API rest.APIHandler `group:"apis"`
}

func newAPIHandler(service *domain.KeySet) apiHandlerResult {
	return apiHandlerResult{
		API: jwksapi.NewRouter(jwksapi.NewHandlers(service)),
	}
}
