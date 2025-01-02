package rdb

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type singleDBManagerTxCtxKey struct{}

type SingleDBManager struct {
	db *sql.DB
}

func NewSingleDBManager(
	userName string,
	password string,
	host string,
	databaseName string,
) (*SingleDBManager, error) {
	// e.g. "postgres://username:password@localhost:5432/database_name"
	uri := "postgres://" + userName + ":" + password + "@" + host + "/" + databaseName + "?sslmode=disable"
	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SingleDBManager{db: db}, nil
}

func (s *SingleDBManager) RunInTx(
	ctx context.Context,
	fn func(context.Context, *sql.Tx) error,
) error {
	return runInTx(ctx, singleDBManagerTxCtxKey{}, s.db, fn)
}

func (s *SingleDBManager) Close() error {
	return s.db.Close()
}
