// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package sqlc

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const bulkUpdateOutboxMessagesAsSent = `-- name: BulkUpdateOutboxMessagesAsSent :exec
UPDATE outbox_messages
SET sent_at = $1
WHERE id = ANY($2::uuid[])
`

type BulkUpdateOutboxMessagesAsSentParams struct {
	SentAt *time.Time
	Ids    []uuid.UUID
}

func (q *Queries) BulkUpdateOutboxMessagesAsSent(ctx context.Context, arg BulkUpdateOutboxMessagesAsSentParams) error {
	_, err := q.db.Exec(ctx, bulkUpdateOutboxMessagesAsSent, arg.SentAt, arg.Ids)
	return err
}

const insertInboxMessage = `-- name: InsertInboxMessage :exec
INSERT INTO inbox_messages (id, payload, received_at, processed_at)
VALUES ($1, $2, $3, $4)
`

type InsertInboxMessageParams struct {
	ID          uuid.UUID
	Payload     []byte
	ReceivedAt  time.Time
	ProcessedAt *time.Time
}

func (q *Queries) InsertInboxMessage(ctx context.Context, arg InsertInboxMessageParams) error {
	_, err := q.db.Exec(ctx, insertInboxMessage,
		arg.ID,
		arg.Payload,
		arg.ReceivedAt,
		arg.ProcessedAt,
	)
	return err
}

const insertOutboxMessage = `-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (id, payload, sent_at)
VALUES ($1, $2, $3)
`

type InsertOutboxMessageParams struct {
	ID      uuid.UUID
	Payload []byte
	SentAt  *time.Time
}

func (q *Queries) InsertOutboxMessage(ctx context.Context, arg InsertOutboxMessageParams) error {
	_, err := q.db.Exec(ctx, insertOutboxMessage, arg.ID, arg.Payload, arg.SentAt)
	return err
}

const selectUnprocessedInboxMessage = `-- name: SelectUnprocessedInboxMessage :one
/**
 * inbox_messages table
 */

SELECT id, payload, received_at, processed_at, created_at, updated_at
FROM inbox_messages
WHERE processed_at IS NULL
ORDER BY received_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

func (q *Queries) SelectUnprocessedInboxMessage(ctx context.Context) (InboxMessage, error) {
	row := q.db.QueryRow(ctx, selectUnprocessedInboxMessage)
	var i InboxMessage
	err := row.Scan(
		&i.ID,
		&i.Payload,
		&i.ReceivedAt,
		&i.ProcessedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const selectUnsentOutboxMessage = `-- name: SelectUnsentOutboxMessage :one
/**
 * outbox_messages table
 */

SELECT id, payload, sent_at, created_at, updated_at
FROM outbox_messages
WHERE sent_at IS NULL
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

func (q *Queries) SelectUnsentOutboxMessage(ctx context.Context) (OutboxMessage, error) {
	row := q.db.QueryRow(ctx, selectUnsentOutboxMessage)
	var i OutboxMessage
	err := row.Scan(
		&i.ID,
		&i.Payload,
		&i.SentAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const selectUnsentOutboxMessages = `-- name: SelectUnsentOutboxMessages :many
SELECT id, payload, sent_at, created_at, updated_at
FROM outbox_messages
WHERE sent_at IS NULL
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED
`

func (q *Queries) SelectUnsentOutboxMessages(ctx context.Context, limit int32) ([]OutboxMessage, error) {
	rows, err := q.db.Query(ctx, selectUnsentOutboxMessages, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []OutboxMessage
	for rows.Next() {
		var i OutboxMessage
		if err := rows.Scan(
			&i.ID,
			&i.Payload,
			&i.SentAt,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateInboxMessage = `-- name: UpdateInboxMessage :exec
UPDATE inbox_messages
SET payload = $2, received_at = $3, processed_at = $4
WHERE id = $1
`

type UpdateInboxMessageParams struct {
	ID          uuid.UUID
	Payload     []byte
	ReceivedAt  time.Time
	ProcessedAt *time.Time
}

func (q *Queries) UpdateInboxMessage(ctx context.Context, arg UpdateInboxMessageParams) error {
	_, err := q.db.Exec(ctx, updateInboxMessage,
		arg.ID,
		arg.Payload,
		arg.ReceivedAt,
		arg.ProcessedAt,
	)
	return err
}

const updateOutboxMessage = `-- name: UpdateOutboxMessage :exec
UPDATE outbox_messages
SET payload = $2, sent_at = $3
WHERE id = $1
`

type UpdateOutboxMessageParams struct {
	ID      uuid.UUID
	Payload []byte
	SentAt  *time.Time
}

func (q *Queries) UpdateOutboxMessage(ctx context.Context, arg UpdateOutboxMessageParams) error {
	_, err := q.db.Exec(ctx, updateOutboxMessage, arg.ID, arg.Payload, arg.SentAt)
	return err
}
