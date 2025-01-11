package sqlc

import (
	"database/sql"

	"github.com/jackc/pgx/v5"
)

func NewDeprecatedQuerier(tx *sql.Tx) Querier {
	return &Queries{db: nil}
}

func NewQuerier(tx pgx.Tx) Querier {
	return &Queries{db: tx}
}
