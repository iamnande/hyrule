package iamjwks

import (
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/domain"
	"github.com/iamnande/hyrule/go/internal/svc/iam-jwks/repository"
)

func newKeySet(repo *repository.EnvRepository) *domain.KeySet {
	return domain.NewKeySet(repo)
}
