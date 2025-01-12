package model

import "github.com/google/uuid"

type InboxMessage struct {
	ID      uuid.UUID
	Payload []byte
}

type InboxMessages []*InboxMessage

func (ms InboxMessages) Filter(excludeIDs []uuid.UUID) InboxMessages {
	var excludeIDMap = make(map[uuid.UUID]struct{}, len(excludeIDs))
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
