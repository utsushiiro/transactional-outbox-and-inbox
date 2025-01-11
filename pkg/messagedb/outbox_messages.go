package messagedb

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/model"
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

func (i *outboxMessages) SelectUnsentOutboxMessage(ctx context.Context) (*model.Message, error) {
	q := i.db.getQuerier(ctx)

	raw, err := q.SelectUnsentOutboxMessage(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrResourceNotFound
		}
		return nil, err
	}

	return &model.Message{
		ID:      raw.MessageUuid,
		Payload: raw.MessagePayload,
	}, nil
}

func (i *outboxMessages) SelectUnsentOutboxMessages(ctx context.Context, size int) (model.Messages, error) {
	q := i.db.getQuerier(ctx)

	raws, err := q.SelectUnsentOutboxMessages(ctx, int32(size))
	if err != nil {
		return nil, err
	}

	msgs := make(model.Messages, 0, len(raws))
	for _, rawMsg := range raws {
		msgs = append(msgs, &model.Message{
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
