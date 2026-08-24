package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/iamnande/hyrule/internal/lib/config"
)

type Params struct {
	fx.In
	Lifecycle fx.Lifecycle

	Logger *slog.Logger
	Config config.Database
}

// NewPool builds the pgx pool every repository shares. pgxpool.New doesn't
// block on connecting, so this pings once on startup - an unreachable
// database fails loud at boot instead of on the first query (see
// docs/conventions.md#probes: this is what startup answers).
func NewPool(params Params) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(params.Config.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	poolConfig.MaxConns = params.Config.MaxConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return fmt.Errorf("ping database: %w", err)
			}
			params.Logger.Info("connected to database", slog.String("host", params.Config.Host))
			return nil
		},
		OnStop: func(_ context.Context) error {
			pool.Close()
			return nil
		},
	})

	return pool, nil
}
