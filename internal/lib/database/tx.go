package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxFn func(ctx context.Context, tx pgx.Tx) error

// WithTx runs fn inside a transaction, committing on success and rolling
// back on error or panic. every repository call goes through this, even a
// single SELECT - RLS policies key off session GUCs (SET LOCAL
// app.account_id = ...) that only live for the current transaction, so this
// is also where that gets set. setGUCs is a no-op today since no policy
// exists yet; the point is that adding one later means touching this one
// function, not every call site.
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
