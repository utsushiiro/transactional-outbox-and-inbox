package messagedb

import (
	"context"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/domain/model"
)

type Transactor interface {
	RunInTx(ctx context.Context, fn func(context.Context) error) error
}

type OutboxMessages interface {
	SelectUnsentOneWithLock(ctx context.Context) (*model.OutboxMessage, error)
	SelectUnsentManyWithLock(ctx context.Context, size int) (model.OutboxMessages, error)
	Insert(ctx context.Context, outboxMessage *model.OutboxMessage) error
	Update(ctx context.Context, outboxMessage *model.OutboxMessage) error
}

type InboxMessages interface {
	SelectUnprocessedOneWithLock(ctx context.Context) (*model.InboxMessage, error)
	Insert(ctx context.Context, inboxMessage *model.InboxMessage) error
	Update(ctx context.Context, inboxMessage *model.InboxMessage) error
}
