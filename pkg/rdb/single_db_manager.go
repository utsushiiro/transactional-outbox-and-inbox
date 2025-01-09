package rdb

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type deprecatedSingleDBManagerTxCtxKey struct{}

type DeprecatedSingleDBManager struct {
	db *sql.DB
}

func NewDeprecatedSingleDBManager(
	userName string,
	password string,
	host string,
	databaseName string,
) (*DeprecatedSingleDBManager, error) {
	// e.g. "postgres://username:password@localhost:5432/database_name"
	uri := "postgres://" + userName + ":" + password + "@" + host + "/" + databaseName + "?sslmode=disable"
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &DeprecatedSingleDBManager{db: db}, nil
}

func (s *DeprecatedSingleDBManager) RunInTx(
	ctx context.Context,
	fn func(context.Context, *sql.Tx) error,
) error {
	return deprecatedRunInTx(ctx, deprecatedSingleDBManagerTxCtxKey{}, s.db, fn)
}

func (s *DeprecatedSingleDBManager) Close() error {
	return s.db.Close()
}

var singleDBManagerTxCtxKey struct{}

type SingleDBManager struct {
	pool *pgxpool.Pool
}

func NewSingleDBManager(
	ctx context.Context,
	userName string,
	password string,
	host string,
	databaseName string,
) (*SingleDBManager, error) {
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

	return &SingleDBManager{pool: pool}, nil
}

func (s *SingleDBManager) RunInTx(
	ctx context.Context,
	fn func(context.Context, pgx.Tx) error,
) error {
	return runInTx(ctx, singleDBManagerTxCtxKey, s.pool, fn)
}

func (s *SingleDBManager) Close() {
	s.pool.Close()
}
