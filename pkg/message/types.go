package message

import "context"

type Message struct {
	ID      string
	Payload []byte
}

type MessageResponder interface {
	Ack()
	Nack()
}

type Publisher interface {
	Publish(ctx context.Context, msg Message) (string, error)
	Close() error
}

type Subscriber interface {
	Receive(ctx context.Context, handler func(context.Context, *Message, MessageResponder)) error
	Close() error
}
