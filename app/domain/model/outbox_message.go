package model

import "github.com/google/uuid"

type OutboxMessage struct {
	ID      uuid.UUID
	Payload []byte
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
