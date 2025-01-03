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
		err := p.publishUnsentMessagesInOutbox(ctx)
		if err != nil {
			log.Printf("failed to publish unsent messages: %v", err)
		}
	}
}

func (p *OutboxWorker) publishUnsentMessagesInOutbox(ctx context.Context) error {
	var count int
	err := p.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)

		unsentMessages, err := querier.SelectUnsentOutboxMessages(ctx, 10)
		if err != nil {
			return err
		}

		for _, unsentMessage := range unsentMessages {
			_, err := p.publisher.Publish(ctx, msgclient.Message{
				ID:      unsentMessage.MessageUuid.String(),
				Payload: []byte(unsentMessage.MessagePayload),
			})
			if err != nil {
				return err
			}
		}

		for _, unsentMessage := range unsentMessages {
			_, err := querier.UpdateOutboxMessageAsSent(ctx, unsentMessage.MessageUuid)
			if err != nil {
				return err
			}

			count++
		}

		return nil
	})
	if err != nil {
		return err
	}

	if count > 0 {
		log.Printf("published %d unsent messages", count)
	} else {
		log.Printf("no unsent messages")
	}

	return nil
}

func (p *OutboxWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
