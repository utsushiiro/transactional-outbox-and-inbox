/**
 * outbox_messages table
 */

-- name: SelectUnsentOutboxMessage :one
SELECT *
FROM outbox_messages
WHERE sent_at IS NULL
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: SelectUnsentOutboxMessages :many
SELECT *
FROM outbox_messages
WHERE sent_at IS NULL
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (id, payload, sent_at)
VALUES ($1, $2, $3);

-- name: UpdateOutboxMessage :exec
UPDATE outbox_messages
SET payload = $2, sent_at = $3
WHERE id = $1;

-- name: BulkUpdateOutboxMessagesAsSent :exec
UPDATE outbox_messages
SET sent_at = $1
WHERE id = ANY(@ids::uuid[]);

/**
 * inbox_messages table
 */

-- name: SelectUnprocessedInboxMessage :one
SELECT *
FROM inbox_messages
WHERE processed_at IS NULL
ORDER BY received_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: InsertInboxMessage :exec
INSERT INTO inbox_messages (id, payload, received_at, processed_at)
VALUES ($1, $2, $3, $4);

-- name: UpdateInboxMessage :exec
UPDATE inbox_messages
SET payload = $2, received_at = $3, processed_at = $4
WHERE id = $1;
