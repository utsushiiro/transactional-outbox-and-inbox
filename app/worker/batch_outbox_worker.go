package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type BatchOutboxWorker struct {
	db                BatchOutboxWorkerMessageDBDeps
	publisher         mq.BatchPublisher
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	batchSize         int
	ticker            *timeutils.Ticker
}

type BatchOutboxWorkerMessageDBDeps struct {
	messagedb.Transactor
	outboxMessages messagedb.OutboxMessages
}

func NewBatchOutboxWorker(
	transactor messagedb.Transactor,
	outboxMessages messagedb.OutboxMessages,
	publisher mq.BatchPublisher,
	poolingInterval time.Duration,
	timeoutPerProcess time.Duration,
	batchSize int,
) *BatchOutboxWorker {
	return &BatchOutboxWorker{
		db: BatchOutboxWorkerMessageDBDeps{
			Transactor:     transactor,
			outboxMessages: outboxMessages,
		},
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
			slog.ErrorContext(ctx, "failed to publish unsent messages", slog.String("error", err.Error()))
		}
	}

	return nil
}

func (p *BatchOutboxWorker) publishUnsentMessagesInOutbox(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeoutPerProcess)
	defer cancel()

	var publishedCount int
	err := p.db.RunInTx(ctx, func(ctx context.Context) error {
		unsentOutboxMessages, err := p.db.outboxMessages.SelectUnsentManyWithLock(ctx, p.batchSize)
		if err != nil {
			return err
		}

		mqMessages := make(mq.Messages, 0, len(unsentOutboxMessages))
		for _, unsentMessage := range unsentOutboxMessages {
			mqMessages = append(mqMessages, &mq.Message{
				ID:      unsentMessage.ID,
				Payload: unsentMessage.Payload,
			})
		}

		result, err := p.publisher.BatchPublish(ctx, mqMessages)
		if err != nil {
			return err
		}

		publishedOutboxMessages := unsentOutboxMessages.Filter(result.FailedIDs)
		publishedCount = len(publishedOutboxMessages)

		sentAt := time.Now()
		err = p.db.outboxMessages.BulkUpdateAsSent(ctx, publishedOutboxMessages.IDs(), sentAt)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if publishedCount > 0 {
		slog.InfoContext(ctx, fmt.Sprintf("published %d unsent messages", publishedCount))
	} else {
		slog.InfoContext(ctx, "no unsent messages")
	}

	return nil
}

func (p *BatchOutboxWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
