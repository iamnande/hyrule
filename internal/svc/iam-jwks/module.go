package iamjwks

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/svc/iam-jwks/domain"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/repository"
)

var Module = fx.Module("iam-jwks",
	fx.Provide(
		repository.New,
		newService,
	),
)

func newService(repo *repository.Repository) *domain.Service {
	return domain.NewService(repo)
}
