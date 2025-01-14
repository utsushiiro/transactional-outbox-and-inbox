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
INSERT INTO outbox_messages (message_uuid, message_payload, sent_at)
VALUES ($1, $2, $3);

-- name: UpdateOutboxMessage :exec
UPDATE outbox_messages
SET message_payload = $2, sent_at = $3
WHERE message_uuid = $1;

-- name: BulkUpdateOutboxMessagesAsSent :exec
UPDATE outbox_messages
SET sent_at = $1
WHERE message_uuid = ANY(@message_uuids::uuid[]);

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
INSERT INTO inbox_messages (message_uuid, message_payload, received_at, processed_at)
VALUES ($1, $2, $3, $4);

-- name: UpdateInboxMessage :exec
UPDATE inbox_messages
SET message_payload = $2, received_at = $3, processed_at = $4
WHERE message_uuid = $1;
