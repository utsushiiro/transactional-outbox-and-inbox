package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type OutboxWorker struct {
	db                OutboxWorkerMessageDBDeps
	publisher         mq.Publisher
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	ticker            *timeutils.Ticker
}

type OutboxWorkerMessageDBDeps struct {
	messagedb.Transactor
	messagedb.OutboxMessages
}

func NewOutboxWorker(
	transactor messagedb.Transactor,
	outboxMessages messagedb.OutboxMessages,
	publisher mq.Publisher,
	poolingInterval time.Duration,
	timeoutPerProcess time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		db: OutboxWorkerMessageDBDeps{
			Transactor:     transactor,
			OutboxMessages: outboxMessages,
		},
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

	err := p.db.RunInTx(ctx, func(ctx context.Context) error {
		unsentMsg, err := p.db.OutboxMessages.SelectUnsentOneWithLock(ctx)
		if err != nil {
			if errors.Is(err, messagedb.ErrResourceNotFound) {
				log.Panicf("no unsent messages")

				return nil
			}

			return err
		}

		err = p.publisher.Publish(ctx, &mq.Message{
			ID:      unsentMsg.ID,
			Payload: unsentMsg.Payload,
		})
		if err != nil {
			return err
		}

		// After publishing the message, mark it as sent.
		unsentMsg.MarkAsSent()

		err = p.db.OutboxMessages.Update(ctx, unsentMsg)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	log.Printf("published an unsent messages")

	return nil
}

func (p *OutboxWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
