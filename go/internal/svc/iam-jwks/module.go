package iamjwks

import (
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/repository"
)

var Module = fx.Module("iam-jwks",
	fx.Provide(
		repository.NewEnv,
		newKeySet,
		newAPIHandler,
	),
)
