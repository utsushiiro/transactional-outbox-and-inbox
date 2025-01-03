// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxMessage struct {
	MessageUuid    uuid.UUID
	MessageTopic   string
	MessagePayload json.RawMessage
	SentAt         sql.NullTime
	CreatedAt      time.Time
	UpdatedAt      time.Time
}