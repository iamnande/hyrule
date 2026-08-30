package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/go/internal/lib/rest/capabilities/health"
)

type HealthCheckResult struct {
	fx.Out

	Option health.Option `group:"healthDependencies"`
}

func NewHealthCheck(pool *pgxpool.Pool) HealthCheckResult {
	return HealthCheckResult{
		Option: health.WithHardDependency("database", pool.Ping),
	}
}
