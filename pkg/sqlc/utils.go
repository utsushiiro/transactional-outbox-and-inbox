package sqlc

import "database/sql"

func NewQuerier(tx *sql.Tx) Querier {
	return &Queries{db: tx}
}
