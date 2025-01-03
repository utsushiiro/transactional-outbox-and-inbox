// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package sqlc

import (
	"context"
	"encoding/json"
)

const insertOutboxMessage = `-- name: InsertOutboxMessage :one
INSERT INTO outbox_messages (message_topic, message_payload)
VALUES ($1, $2)
RETURNING message_uuid, message_topic, message_payload, sent_at, created_at, updated_at
`

type InsertOutboxMessageParams struct {
	MessageTopic   string
	MessagePayload json.RawMessage
}

func (q *Queries) InsertOutboxMessage(ctx context.Context, arg InsertOutboxMessageParams) (OutboxMessage, error) {
	row := q.db.QueryRowContext(ctx, insertOutboxMessage, arg.MessageTopic, arg.MessagePayload)
	var i OutboxMessage
	err := row.Scan(
		&i.MessageUuid,
		&i.MessageTopic,
		&i.MessagePayload,
		&i.SentAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}