package messagedb

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/domain/model"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/messagedb/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/telemetry"
)

type OutboxMessages struct {
	db *DB
}

var _ messagedb.OutboxMessages = (*OutboxMessages)(nil)

func NewOutboxMessages(
	db *DB,
) *OutboxMessages {
	return &OutboxMessages{
		db: db,
	}
}

func (o *OutboxMessages) SelectUnsentOneWithLock(ctx context.Context) (*model.OutboxMessage, error) {
	ctx, span := telemetry.StartSpanWithFuncName(ctx)
	defer span.End()

	q := o.db.getQuerier(ctx)

	raw, err := q.SelectUnsentOutboxMessage(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, messagedb.ErrResourceNotFound
		}

		return nil, err
	}

	return &model.OutboxMessage{
		ID:      raw.ID,
		Payload: raw.Payload,
		SentAt:  raw.SentAt,
	}, nil
}

func (o *OutboxMessages) SelectUnsentManyWithLock(ctx context.Context, size int) (model.OutboxMessages, error) {
	ctx, span := telemetry.StartSpanWithFuncName(ctx)
	defer span.End()

	q := o.db.getQuerier(ctx)

	if size > math.MaxInt32 {
		return nil, fmt.Errorf("size must be less than %d", math.MaxInt32)
	}

	//nolint:gosec // size is validated above
	raws, err := q.SelectUnsentOutboxMessages(ctx, int32(size))
	if err != nil {
		return nil, err
	}

	outboxMessages := make(model.OutboxMessages, len(raws))
	for i, raw := range raws {
		outboxMessages[i] = &model.OutboxMessage{
			ID:      raw.ID,
			Payload: raw.Payload,
			SentAt:  raw.SentAt,
		}
	}

	return outboxMessages, nil
}

func (o *OutboxMessages) Insert(ctx context.Context, outboxMessage *model.OutboxMessage) error {
	ctx, span := telemetry.StartSpanWithFuncName(ctx)
	defer span.End()

	q := o.db.getQuerier(ctx)

	err := q.InsertOutboxMessage(ctx, sqlc.InsertOutboxMessageParams{
		ID:      outboxMessage.ID,
		Payload: outboxMessage.Payload,
		SentAt:  outboxMessage.SentAt,
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *OutboxMessages) Update(ctx context.Context, outboxMessage *model.OutboxMessage) error {
	ctx, span := telemetry.StartSpanWithFuncName(ctx)
	defer span.End()

	q := o.db.getQuerier(ctx)

	err := q.UpdateOutboxMessage(ctx, sqlc.UpdateOutboxMessageParams{
		ID:      outboxMessage.ID,
		Payload: outboxMessage.Payload,
		SentAt:  outboxMessage.SentAt,
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *OutboxMessages) BulkUpdateAsSent(ctx context.Context, outboxMessageIDs []uuid.UUID, sentAt time.Time) error {
	ctx, span := telemetry.StartSpanWithFuncName(ctx)
	defer span.End()

	q := o.db.getQuerier(ctx)

	err := q.BulkUpdateOutboxMessagesAsSent(ctx, sqlc.BulkUpdateOutboxMessagesAsSentParams{
		Ids:    outboxMessageIDs,
		SentAt: &sentAt,
	})
	if err != nil {
		return err
	}

	return nil
}
