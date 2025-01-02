package rdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrTxAlreadyStarted = errors.New("transaction already started in passed context")

func runInTx(
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
