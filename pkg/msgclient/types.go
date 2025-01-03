package msgclient

type Message struct {
	ID      string
	Payload []byte
}

type MessageResponder interface {
	Ack()
	Nack()
}
