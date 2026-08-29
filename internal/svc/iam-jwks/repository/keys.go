package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iamnande/hyrule/internal/lib/database"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/domain"
	"github.com/iamnande/hyrule/internal/svc/iam-jwks/repository/key"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context) ([]domain.Key, error) {
	var keys []domain.Key
	err := database.WithTx(ctx, r.pool, func(ctx context.Context, tx pgx.Tx) error {
		rows, err := key.New(tx).List(ctx)
		if err != nil {
			return fmt.Errorf("list keys: %w", err)
		}
		keys = make([]domain.Key, len(rows))
		for i, row := range rows {
			keys[i] = toDomain(row)
		}
		return nil
	})
	return keys, err
}

func toDomain(row key.IamJwksKey) domain.Key {
	return domain.Key{
		ID:        row.ID,
		Algorithm: row.Algorithm,
		PublicKey: row.PublicKey,
		CreatedAt: row.CreatedAt.Time,
	}
}
