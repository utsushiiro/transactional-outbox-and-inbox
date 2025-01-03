-- name: InsertOutboxMessage :one
INSERT INTO outbox_messages (message_topic, message_payload)
VALUES ($1, $2)
RETURNING *;

-- name: SelectUnsentOutboxMessages :many
SELECT *
FROM outbox_messages
WHERE sent_at IS NULL
ORDER BY created_at ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: UpdateOutboxMessageAsSent :one
UPDATE outbox_messages
SET sent_at = NOW()
WHERE message_uuid = $1
RETURNING *;

-- name: InsertInboxMessage :one
INSERT INTO inbox_messages (message_uuid, message_payload, received_at)
VALUES ($1, $2, NOW())
RETURNING *;
