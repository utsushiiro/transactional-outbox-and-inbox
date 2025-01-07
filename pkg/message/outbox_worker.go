package message

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
)

type OutboxWorker struct {
	dbManager         *rdb.SingleDBManager
	publisher         Publisher
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	ticker            *timeutils.Ticker
}

type Publisher interface {
	Publish(ctx context.Context, msg *Message) error
	Close() error
}

type BatchPublisher interface {
	BatchPublish(ctx context.Context, msgs []*Message) (*BatchResult, error)
	Close() error
}

type BatchResult struct {
	SucceededIDs []string
	FailedIDs    []string
}

func NewOutboxWorker(
	dbManager *rdb.SingleDBManager,
	publisher Publisher,
	poolingInterval time.Duration,
	timeoutPerProcess time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		dbManager:         dbManager,
		publisher:         publisher,
		pollingInterval:   poolingInterval,
		timeoutPerProcess: timeoutPerProcess,
	}
}

func (p *OutboxWorker) Run() error {
	ctx := context.Background()
	ticker := timeutils.NewTicker(p.pollingInterval)
	p.ticker = ticker

	for range ticker.C() {
		err := p.publishUnsentMessagesInOutbox(ctx)
		if err != nil {
			log.Printf("failed to publish unsent messages: %v", err)
		}
	}

	return nil
}

func (p *OutboxWorker) publishUnsentMessagesInOutbox(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeoutPerProcess)
	defer cancel()

	var count int
	err := p.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)

		unsentMessages, err := querier.SelectUnsentOutboxMessages(ctx, 10)
		if err != nil {
			return err
		}

		for _, unsentMessage := range unsentMessages {
			err := p.publisher.Publish(ctx, &Message{
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
