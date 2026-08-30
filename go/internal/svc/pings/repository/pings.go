package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iamnande/hyrule/go/internal/lib/database"
	"github.com/iamnande/hyrule/go/internal/svc/pings/domain"
	"github.com/iamnande/hyrule/go/internal/svc/pings/repository/ping"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Upsert(ctx context.Context, name string, kind domain.Kind) (domain.Ping, error) {
	var result domain.Ping
	err := database.WithTx(ctx, r.pool, func(ctx context.Context, tx pgx.Tx) error {
		row, err := ping.New(tx).Upsert(ctx, ping.UpsertParams{
			Name: name,
			Kind: string(kind),
		})
		if err != nil {
			return fmt.Errorf("upsert ping: %w", err)
		}
		result = toDomain(row)
		return nil
	})
	return result, err
}

func (r *Repository) List(ctx context.Context) ([]domain.Ping, error) {
	var pings []domain.Ping
	err := database.WithTx(ctx, r.pool, func(ctx context.Context, tx pgx.Tx) error {
		rows, err := ping.New(tx).List(ctx)
		if err != nil {
			return fmt.Errorf("list pings: %w", err)
		}
		pings = make([]domain.Ping, len(rows))
		for i, row := range rows {
			pings[i] = toDomain(row)
		}
		return nil
	})
	return pings, err
}

func toDomain(row ping.Ping) domain.Ping {
	return domain.Ping{
		Name:        row.Name,
		Kind:        domain.Kind(row.Kind),
		FirstSeenAt: row.FirstSeenAt.Time,
		LastSeenAt:  row.LastSeenAt.Time,
	}
}
