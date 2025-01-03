package message

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/msgclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
)

type OutboxWorker struct {
	dbManager *rdb.SingleDBManager
	publisher msgclient.Publisher
	ticker    *timeutils.Ticker
}

func NewOutboxWorker(dbManager *rdb.SingleDBManager, publisher msgclient.Publisher) *OutboxWorker {
	return &OutboxWorker{
		dbManager: dbManager,
		publisher: publisher,
	}
}

func (p *OutboxWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := timeutils.NewTicker(interval)
	p.ticker = ticker

	for range ticker.C() {
		p.publishUnsentMessagesInOutbox(ctx)
	}
}

func (p *OutboxWorker) publishUnsentMessagesInOutbox(ctx context.Context) {
	_ = p.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)

		unsentMessages, err := querier.SelectUnsentOutboxMessages(ctx, 10)
		if err != nil {
			return err
		}

		for _, unsentMessage := range unsentMessages {
			msgID, err := p.publisher.Publish(ctx, msgclient.Message{
				ID:      unsentMessage.MessageUuid.String(),
				Payload: []byte(unsentMessage.MessagePayload),
			})
			if err != nil {
				return err
			}

			log.Printf("Published message: %s", msgID)
		}

		for _, unsentMessage := range unsentMessages {
			updated, err := querier.UpdateOutboxMessageAsSent(ctx, unsentMessage.MessageUuid)
			if err != nil {
				return err
			}

			log.Printf("Updated messageUuid: %s", updated.MessageUuid)
		}

		return nil
	})
}

func (p *OutboxWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
