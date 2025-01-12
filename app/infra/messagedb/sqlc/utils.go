package sqlc

import (
	"github.com/jackc/pgx/v5"
)

func NewQuerier(tx pgx.Tx) Querier {
	return &Queries{db: tx}
}
