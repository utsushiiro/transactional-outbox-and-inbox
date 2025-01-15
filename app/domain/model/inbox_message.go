package model

import (
	"time"

	"github.com/google/uuid"

	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type InboxMessage struct {
	ID          uuid.UUID
	Payload     []byte
	ReceivedAt  time.Time
	ProcessedAt *time.Time
}

func NewInboxMessage(id uuid.UUID, payload []byte) *InboxMessage {
	return &InboxMessage{
		ID:          id,
		Payload:     payload,
		ReceivedAt:  timeutils.NowUTC(),
		ProcessedAt: nil,
	}
}

func (i *InboxMessage) MarkAsProcessed() {
	now := timeutils.NowUTC()
	i.ProcessedAt = &now
}

type InboxMessages []*InboxMessage

func (ms InboxMessages) Filter(excludeIDs []uuid.UUID) InboxMessages {
	excludeIDMap := make(map[uuid.UUID]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		excludeIDMap[id] = struct{}{}
	}

	var filteredMessages InboxMessages
	for _, m := range ms {
		if _, ok := excludeIDMap[m.ID]; !ok {
			filteredMessages = append(filteredMessages, m)
		}
	}

	return filteredMessages
}
