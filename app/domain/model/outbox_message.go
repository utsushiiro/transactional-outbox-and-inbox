package model

import (
	"time"

	"github.com/google/uuid"
)

type OutboxMessage struct {
	ID      uuid.UUID
	Payload []byte
	SentAt  *time.Time
}

func NewOutboxMessage(payload []byte) *OutboxMessage {
	return &OutboxMessage{
		ID:      uuid.New(),
		Payload: payload,
	}
}

func (m *OutboxMessage) MarkAsSent() {
	now := time.Now()
	m.SentAt = &now
}

type OutboxMessages []*OutboxMessage

func (ms OutboxMessages) Filter(excludeIDs []uuid.UUID) OutboxMessages {
	var excludeIDMap = make(map[uuid.UUID]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		excludeIDMap[id] = struct{}{}
	}

	var filteredMessages OutboxMessages
	for _, m := range ms {
		if _, ok := excludeIDMap[m.ID]; !ok {
			filteredMessages = append(filteredMessages, m)
		}
	}

	return filteredMessages
}

func (ms OutboxMessages) IDs() []uuid.UUID {
	ids := make([]uuid.UUID, len(ms))
	for i, m := range ms {
		ids[i] = m.ID
	}

	return ids
}