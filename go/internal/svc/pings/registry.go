package pings

import (
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
	"github.com/iamnande/hyrule/go/internal/svc/pings/repository"
)

func newRegistry(repo *repository.PostgresRepository, cfg domain.Config) *domain.Registry {
	return domain.NewRegistry(repo, cfg)
}
