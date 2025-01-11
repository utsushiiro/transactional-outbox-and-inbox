package messagedb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
)

type DB struct {
	pool *pgxpool.Pool

	*inboxMessages
	*outboxMessages
}

func NewDB(
	ctx context.Context,
	userName string,
	password string,
	host string,
	databaseName string,
) (*DB, error) {
	// e.g. "postgres://username:password@localhost:5432/database_name"
	uri := "postgres://" + userName + ":" + password + "@" + host + "/" + databaseName + "?sslmode=disable"
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	db := &DB{
		pool: pool,
	}

	db.inboxMessages = newInboxMessages(db)
	db.outboxMessages = newOutboxMessages(db)

	return db, nil
}

var ErrTxAlreadyStarted = errors.New("transaction already started in passed context")

type ctxKeyForPgxTx struct{}

func (d *DB) RunInTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	_, ok := ctx.Value(ctxKeyForPgxTx{}).(pgx.Tx)
	if ok {
		return ErrTxAlreadyStarted
	}

	tx, err := d.pool.Begin(ctx)
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

	ctxWithTx := context.WithValue(ctx, ctxKeyForPgxTx{}, tx)

	if fnErr := fn(ctxWithTx); fnErr != nil {
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

func (d *DB) getQuerier(ctx context.Context) sqlc.Querier {
	if v, ok := ctx.Value(ctxKeyForPgxTx{}).(pgx.Tx); ok {
		return sqlc.NewQuerier(v)
	}

	return sqlc.New(d.pool)
}
