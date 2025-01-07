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
	Publish(ctx context.Context, msg *Message) error
	Close() error
}

type BatchPublisher interface {
	BatchPublish(ctx context.Context, msgs []*Message) (*BatchResult, error)
	Close() error
}

type BatchResult struct {
	SucceededIDs []string
	FailedIDs    []string
}

type Subscriber interface {
	Receive(ctx context.Context, handler func(context.Context, *Message, MessageResponder)) error
	Close() error
}
