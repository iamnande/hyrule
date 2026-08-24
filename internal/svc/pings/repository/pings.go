package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iamnande/hyrule/internal/lib/database"
	"github.com/iamnande/hyrule/internal/svc/pings/domain"
	generated "github.com/iamnande/hyrule/internal/svc/pings/repository/generated"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Upsert(ctx context.Context, name string, kind domain.Kind) (domain.Ping, error) {
	var ping domain.Ping
	err := database.WithTx(ctx, r.pool, func(ctx context.Context, tx pgx.Tx) error {
		row, err := generated.New(tx).UpsertPing(ctx, generated.UpsertPingParams{
			Name: name,
			Kind: string(kind),
		})
		if err != nil {
			return fmt.Errorf("upsert ping: %w", err)
		}
		ping = toDomain(row)
		return nil
	})
	return ping, err
}

func (r *Repository) List(ctx context.Context) ([]domain.Ping, error) {
	var pings []domain.Ping
	err := database.WithTx(ctx, r.pool, func(ctx context.Context, tx pgx.Tx) error {
		rows, err := generated.New(tx).ListPings(ctx)
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

func toDomain(row generated.Ping) domain.Ping {
	return domain.Ping{
		Name:        row.Name,
		Kind:        domain.Kind(row.Kind),
		FirstSeenAt: row.FirstSeenAt.Time,
		LastSeenAt:  row.LastSeenAt.Time,
	}
}
