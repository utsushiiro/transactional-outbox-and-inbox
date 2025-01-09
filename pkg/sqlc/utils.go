package sqlc

import "database/sql"

func NewDeprecatedQuerier(tx *sql.Tx) Querier {
	return &Queries{db: tx}
}
