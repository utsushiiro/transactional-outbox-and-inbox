package message

import "github.com/google/uuid"

type Message struct {
	ID      uuid.UUID
	Payload []byte
}

type Messages []*Message

func (ms Messages) Filter(excludeIDs []uuid.UUID) Messages {
	var excludeIDMap = make(map[uuid.UUID]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		excludeIDMap[id] = struct{}{}
	}

	var filteredMessages Messages
	for _, m := range ms {
		if _, ok := excludeIDMap[m.ID]; !ok {
			filteredMessages = append(filteredMessages, m)
		}
	}

	return filteredMessages
}
