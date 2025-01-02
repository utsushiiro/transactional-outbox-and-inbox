-- name: InsertOutboxMessage :one
INSERT INTO outbox_messages (message_topic, message_payload)
VALUES ($1, $2)
RETURNING *;
