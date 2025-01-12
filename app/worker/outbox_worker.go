package worker

import (
	"context"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type OutboxWorker struct {
	db                *messagedb.DB
	publisher         mq.Publisher
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	ticker            *timeutils.Ticker
}

func NewOutboxWorker(
	db *messagedb.DB,
	publisher mq.Publisher,
	poolingInterval time.Duration,
	timeoutPerProcess time.Duration,
) *OutboxWorker {
	return &OutboxWorker{
		db:                db,
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

	var isSent bool
	err := p.db.RunInTx(ctx, func(ctx context.Context) error {
		unsentMsg, err := p.db.SelectUnsentOutboxMessage(ctx)
		if err != nil {
			return err
		}

		err = p.publisher.Publish(ctx, unsentMsg)
		if err != nil {
			return err
		}

		err = p.db.UpdateOutboxMessageAsSent(ctx, unsentMsg.ID)
		if err != nil {
			return err
		}

		isSent = true

		return nil
	})
	if err != nil {
		return err
	}

	if isSent {
		log.Printf("published an unsent messages")
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
