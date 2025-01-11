package message

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/model"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type BatchOutboxWorker struct {
	db                *messagedb.DB
	publisher         BatchPublisher
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	batchSize         int
	ticker            *timeutils.Ticker
}

type BatchPublisher interface {
	BatchPublish(ctx context.Context, msgs []*model.Message) (*BatchResult, error)
	Close() error
}

type BatchResult struct {
	SucceededIDs []uuid.UUID
	FailedIDs    []uuid.UUID
}

func NewBatchOutboxWorker(
	db *messagedb.DB,
	publisher BatchPublisher,
	poolingInterval time.Duration,
	timeoutPerProcess time.Duration,
	batchSize int,
) *BatchOutboxWorker {
	return &BatchOutboxWorker{
		db:                db,
		publisher:         publisher,
		pollingInterval:   poolingInterval,
		timeoutPerProcess: timeoutPerProcess,
		batchSize:         batchSize,
	}
}

func (p *BatchOutboxWorker) Run() error {
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

func (p *BatchOutboxWorker) publishUnsentMessagesInOutbox(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeoutPerProcess)
	defer cancel()

	var publishedCount int
	err := p.db.RunInTx(ctx, func(ctx context.Context) error {
		unsentMsgs, err := p.db.SelectUnsentOutboxMessages(ctx, p.batchSize)
		if err != nil {
			return err
		}

		result, err := p.publisher.BatchPublish(ctx, unsentMsgs)
		if err != nil {
			return err
		}

		publishedMsgs := unsentMsgs.Filter(result.FailedIDs)
		publishedCount = len(publishedMsgs)

		// TODO: use bulk update
		var errs error
		for _, publishedMsg := range publishedMsgs {
			err = p.db.UpdateOutboxMessageAsSent(ctx, publishedMsg.ID)
			if err != nil {
				errs = errors.Join(errs, err)
			}
		}
		if errs != nil {
			return errs
		}

		return nil
	})
	if err != nil {
		return err
	}

	if publishedCount > 0 {
		log.Printf("published %d unsent messages", publishedCount)
	} else {
		log.Printf("no unsent messages")
	}

	return nil
}

func (p *BatchOutboxWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
