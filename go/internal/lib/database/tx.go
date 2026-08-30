package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxFn func(ctx context.Context, tx pgx.Tx) error

func WithTx(ctx context.Context, pool *pgxpool.Pool, fn TxFn) (err error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	if err = setGUCs(ctx, tx); err != nil {
		return fmt.Errorf("set session GUCs: %w", err)
	}

	return fn(ctx, tx)
}

func setGUCs(_ context.Context, _ pgx.Tx) error {
	return nil
}
