package registration

import (
	"context"

	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/config"
	"github.com/iamnande/hyrule/internal/database"
)

type DatabaseClient interface {
	database.TransactWriteItems
}

type PasswordService interface {
	HashPassword(ctx context.Context, password string) (string, error)
}

type Service struct {
	databaseClient DatabaseClient
	databaseConfig config.Database

	passwordService PasswordService
}

type Params struct {
	fx.In

	DatabaseClient DatabaseClient
	DatabaseConfig config.Database

	PasswordService PasswordService
}

func NewService(params Params) (*Service, error) {
	return &Service{
		databaseClient:  params.DatabaseClient,
		databaseConfig:  params.DatabaseConfig,
		passwordService: params.PasswordService,
	}, nil
}
