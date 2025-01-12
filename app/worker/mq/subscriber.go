package mq

import (
	"context"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/model"
)

type Subscriber interface {
	Receive(ctx context.Context, handler func(context.Context, *model.Message, MessageResponder)) error
	Close() error
}

type MessageResponder interface {
	Ack()
	Nack()
}
