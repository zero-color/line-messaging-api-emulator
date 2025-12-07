package db

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5"
)

func RunInTx(ctx context.Context, f func(ctx context.Context) error) error {
	ctxDB := FromContext(ctx)
	if _, ok := ctxDB.(*txDBClient); ok {
		return f(ctx)
	}

	db, ok := ctxDB.(*dbClient)
	if !ok {
		// This is only for testing purposes, where we use mock or other implementations
		return f(ctx)
	}

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}

	ctx = ToContext(ctx, newTxClient(tx))
	if err := f(ctx); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return errors.Wrap(err, "rollback transaction")
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "commit transaction")
	}
	return nil
}
