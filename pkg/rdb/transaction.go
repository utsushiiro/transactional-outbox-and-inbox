package rdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTxAlreadyStarted = errors.New("transaction already started in passed context")

func deprecatedRunInTx(
	ctx context.Context,
	ctxKeyForTx any,
	db *sql.DB,
	fn func(context.Context, *sql.Tx) error,
) error {
	_, ok := ctx.Value(ctxKeyForTx).(*sql.Tx)
	if ok {
		return ErrTxAlreadyStarted
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			var panicErr error
			switch x := p.(type) {
			case error:
				panicErr = x
			default:
				panicErr = fmt.Errorf("%v", x)
			}

			if rbErr := tx.Rollback(); rbErr != nil {
				panic(errors.Join(panicErr, rbErr))
			}

			panic(panicErr)
		}
	}()

	ctxWithTx := context.WithValue(ctx, ctxKeyForTx, tx)

	if fnErr := fn(ctxWithTx, tx); fnErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(fnErr, rbErr)
		}

		return fnErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(commitErr, rbErr)
		}

		return commitErr
	}

	return nil
}

func runInTx(
	ctx context.Context,
	ctxKeyForTx any,
	db *pgxpool.Pool,
	fn func(context.Context, pgx.Tx) error,
) error {
	_, ok := ctx.Value(ctxKeyForTx).(*sql.Tx)
	if ok {
		return ErrTxAlreadyStarted
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			var panicErr error
			switch x := p.(type) {
			case error:
				panicErr = x
			default:
				panicErr = fmt.Errorf("%v", x)
			}

			if rbErr := tx.Rollback(ctx); rbErr != nil {
				panic(errors.Join(panicErr, rbErr))
			}

			panic(panicErr)
		}
	}()

	ctxWithTx := context.WithValue(ctx, ctxKeyForTx, tx)

	if fnErr := fn(ctxWithTx, tx); fnErr != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(fnErr, rbErr)
		}

		return fnErr
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.Join(commitErr, rbErr)
		}

		return commitErr
	}

	return nil
}
