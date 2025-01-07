package message

type Message struct {
	ID      string
	Payload []byte
}

type Messages []*Message

func (ms Messages) Filter(excludeIDs []string) Messages {
	var excludeIDMap = make(map[string]struct{}, len(excludeIDs))
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
