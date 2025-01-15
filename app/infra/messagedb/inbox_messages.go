package messagedb

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/domain/model"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/messagedb/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/messagedb"
)

type InboxMessages struct {
	db *DB
}

var _ messagedb.InboxMessages = &InboxMessages{}

func NewInboxMessages(
	db *DB,
) *InboxMessages {
	return &InboxMessages{
		db: db,
	}
}

func (i *InboxMessages) SelectUnprocessedOneWithLock(ctx context.Context) (*model.InboxMessage, error) {
	q := i.db.getQuerier(ctx)

	raw, err := q.SelectUnprocessedInboxMessage(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, messagedb.ErrResourceNotFound
		}

		return nil, err
	}

	return &model.InboxMessage{
		ID:          raw.MessageUuid,
		Payload:     raw.MessagePayload,
		ReceivedAt:  raw.ReceivedAt,
		ProcessedAt: raw.ProcessedAt,
	}, nil
}

func (i *InboxMessages) Insert(ctx context.Context, inboxMessage *model.InboxMessage) error {
	q := i.db.getQuerier(ctx)

	err := q.InsertInboxMessage(ctx, sqlc.InsertInboxMessageParams{
		MessageUuid:    inboxMessage.ID,
		MessagePayload: inboxMessage.Payload,
		ReceivedAt:     inboxMessage.ReceivedAt,
		ProcessedAt:    inboxMessage.ProcessedAt,
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *InboxMessages) Update(ctx context.Context, inboxMessage *model.InboxMessage) error {
	q := i.db.getQuerier(ctx)

	err := q.UpdateInboxMessage(ctx, sqlc.UpdateInboxMessageParams{
		MessageUuid:    inboxMessage.ID,
		MessagePayload: inboxMessage.Payload,
		ReceivedAt:     inboxMessage.ReceivedAt,
		ProcessedAt:    inboxMessage.ProcessedAt,
	})
	if err != nil {
		return err
	}

	return nil
}
