package rdb

import (
	"context"
	"database/sql"
)

type singleDBManagerTxCtxKey struct{}

type SingleDBManager struct {
	db *sql.DB
}

func NewSingleDBManager(db *sql.DB) *SingleDBManager {
	return &SingleDBManager{db: db}
}

func (s *SingleDBManager) RunInTx(
	ctx context.Context,
	fn func(context.Context, *sql.Tx) error,
) error {
	return runInTx(ctx, singleDBManagerTxCtxKey{}, s.db, fn)
}
