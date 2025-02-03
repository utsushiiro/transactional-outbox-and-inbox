package messagedb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/metric"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/messagedb/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/telemetry"
)

type DB struct {
	pool *pgxpool.Pool
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

	config.MaxConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnLifetimeJitter = config.MaxConnLifetime / 2
	// All connections live for a maximum amount of time even if they are not in use.
	config.MaxConnIdleTime = config.MaxConnLifetime + config.MaxConnLifetimeJitter

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	db := &DB{
		pool: pool,
	}

	totalConns, err := telemetry.Meter.Int64ObservableGauge(
		"pgxpool_total_connections",
		metric.WithDescription("The total number of connections in the pgx connection pool, including both idle and active connections."),
	)
	if err != nil {
		return nil, err
	}

	idleConns, err := telemetry.Meter.Int64ObservableGauge(
		"pgxpool_idle_connections",
		metric.WithDescription("The number of idle connections in the pgx connection pool that are available for use."),
	)
	if err != nil {
		return nil, err
	}

	activeConns, err := telemetry.Meter.Int64ObservableGauge(
		"pgxpool_active_connections",
		metric.WithDescription("The number of active connections currently being used from the pgx connection pool."),
	)
	if err != nil {
		return nil, err
	}

	waitCount, err := telemetry.Meter.Int64ObservableGauge(
		"pgxpool_wait_count",
		metric.WithDescription("The total number of times a connection has been requested from the pgx connection pool."),
	)
	if err != nil {
		return nil, err
	}

	waitDuration, err := telemetry.Meter.Float64ObservableGauge(
		"pgxpool_wait_duration_seconds",
		metric.WithDescription("The total duration in seconds spent waiting for a connection from the pgx connection pool."),
	)
	if err != nil {
		return nil, err
	}
	_, err = telemetry.Meter.RegisterCallback(
		func(_ context.Context, observer metric.Observer) error {
			poolStats := pool.Stat()

			observer.ObserveInt64(totalConns, int64(poolStats.TotalConns()))
			observer.ObserveInt64(idleConns, int64(poolStats.IdleConns()))
			observer.ObserveInt64(activeConns, int64(poolStats.AcquiredConns()))
			observer.ObserveInt64(waitCount, poolStats.AcquireCount())
			observer.ObserveFloat64(waitDuration, poolStats.AcquireDuration().Seconds())

			return nil
		},
		totalConns, idleConns, activeConns, waitCount, waitDuration,
	)
	if err != nil {
		return nil, err
	}

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
		return sqlc.New(v)
	}

	return sqlc.New(d.pool)
}
