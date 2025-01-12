package messagedb

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/messagedb/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/model"
)

type inboxMessages struct {
	db *DB
}

func newInboxMessages(db *DB) *inboxMessages {
	return &inboxMessages{
		db: db,
	}
}

type InsertInboxMessageParams struct {
	MessageID uuid.UUID
	Payload   []byte
}

func (i *inboxMessages) InsertInboxMessage(ctx context.Context, param *InsertInboxMessageParams) error {
	q := i.db.getQuerier(ctx)

	_, err := q.InsertInboxMessage(ctx, sqlc.InsertInboxMessageParams{
		MessageUuid:    param.MessageID,
		MessagePayload: param.Payload,
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *inboxMessages) SelectUnprocessedInboxMessage(ctx context.Context) (*model.Message, error) {
	q := i.db.getQuerier(ctx)

	raw, err := q.SelectUnprocessedInboxMessage(ctx)
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

func (i *inboxMessages) UpdateInboxMessageAsProcessed(ctx context.Context, messageID uuid.UUID) error {
	q := i.db.getQuerier(ctx)

	_, err := q.UpdateInboxMessageAsProcessed(ctx, messageID)
	if err != nil {
		return err
	}

	return nil
}
