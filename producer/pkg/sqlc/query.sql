-- name: GetMessageByID :one
SELECT * FROM message_outbox
WHERE id = $1 LIMIT 1;
