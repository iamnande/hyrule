package user

import (
	"github.com/iamnande/hyrule/internal/config"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/database"
)

type DatabaseClient interface {
	database.PutItem
}

type Repository struct {
	db  DatabaseClient
	cfg config.Database
}

type Params struct {
	fx.In

	DatabaseClient DatabaseClient
	DatabaseConfig config.Database
}

func NewRepository(params Params) *Repository {
	return &Repository{
		db:  params.DatabaseClient,
		cfg: params.DatabaseConfig,
	}
}
