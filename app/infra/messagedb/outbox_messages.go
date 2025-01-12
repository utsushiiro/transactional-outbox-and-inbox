package messagedb

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/domain/model"
)

type outboxMessages struct {
	db *DB
}

func newOutboxMessages(db *DB) *outboxMessages {
	return &outboxMessages{
		db: db,
	}
}

func (i *outboxMessages) InsertOutboxMessage(ctx context.Context, messagePayload []byte) error {
	q := i.db.getQuerier(ctx)

	_, err := q.InsertOutboxMessage(ctx, messagePayload)
	if err != nil {
		return err
	}

	return nil
}

func (i *outboxMessages) SelectUnsentOutboxMessage(ctx context.Context) (*model.OutboxMessage, error) {
	q := i.db.getQuerier(ctx)

	raw, err := q.SelectUnsentOutboxMessage(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return &model.OutboxMessage{
		ID:      raw.MessageUuid,
		Payload: raw.MessagePayload,
	}, nil
}

func (i *outboxMessages) SelectUnsentOutboxMessages(ctx context.Context, size int) (model.OutboxMessages, error) {
	q := i.db.getQuerier(ctx)

	raws, err := q.SelectUnsentOutboxMessages(ctx, int32(size))
	if err != nil {
		return nil, err
	}

	msgs := make(model.OutboxMessages, 0, len(raws))
	for _, rawMsg := range raws {
		msgs = append(msgs, &model.OutboxMessage{
			ID:      rawMsg.MessageUuid,
			Payload: rawMsg.MessagePayload,
		})
	}

	return msgs, nil
}

func (i *outboxMessages) UpdateOutboxMessageAsSent(ctx context.Context, messageID uuid.UUID) error {
	q := i.db.getQuerier(ctx)

	_, err := q.UpdateOutboxMessageAsSent(ctx, messageID)
	if err != nil {
		return err
	}

	return nil
}
