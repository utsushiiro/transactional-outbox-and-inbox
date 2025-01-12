package mq

import (
	"context"
)

type Subscriber interface {
	Receive(ctx context.Context, handler func(context.Context, *Message, MessageResponder)) error
	Close() error
}

type MessageResponder interface {
	Ack()
	Nack()
}
