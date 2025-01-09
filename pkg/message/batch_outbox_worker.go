package message

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
)

type BatchOutboxWorker struct {
	dbManager         *rdb.DeprecatedSingleDBManager
	publisher         BatchPublisher
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	batchSize         int
	ticker            *timeutils.Ticker
}

type BatchPublisher interface {
	BatchPublish(ctx context.Context, msgs []*Message) (*BatchResult, error)
	Close() error
}

type BatchResult struct {
	SucceededIDs []uuid.UUID
	FailedIDs    []uuid.UUID
}

func NewBatchOutboxWorker(
	dbManager *rdb.DeprecatedSingleDBManager,
	publisher BatchPublisher,
	poolingInterval time.Duration,
	timeoutPerProcess time.Duration,
	batchSize int,
) *BatchOutboxWorker {
	return &BatchOutboxWorker{
		dbManager:         dbManager,
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
	err := p.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewDeprecatedQuerier(tx)

		unsentMessages, err := SelectUnsentOutboxMessages(ctx, querier, int32(p.batchSize))
		if err != nil {
			return err
		}

		result, err := p.publisher.BatchPublish(ctx, unsentMessages)
		if err != nil {
			return err
		}

		publishedMsgs := unsentMessages.Filter(result.FailedIDs)
		publishedCount = len(publishedMsgs)

		err = UpdateOutboxMessagesAsSent(ctx, querier, publishedMsgs)
		if err != nil {
			return err
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

func SelectUnsentOutboxMessages(ctx context.Context, querier sqlc.Querier, limit int32) (Messages, error) {
	unsentMessages, err := querier.SelectUnsentOutboxMessages(ctx, limit)
	if err != nil {
		return nil, err
	}

	msgs := make(Messages, 0, len(unsentMessages))
	for _, unsentMessage := range unsentMessages {
		msgs = append(msgs, &Message{
			ID:      unsentMessage.MessageUuid,
			Payload: []byte(unsentMessage.MessagePayload),
		})
	}

	return msgs, nil
}

// TODO: use bulk update
func UpdateOutboxMessagesAsSent(ctx context.Context, querier sqlc.Querier, publishedMessages Messages) error {
	for _, publishedMessage := range publishedMessages {
		_, err := querier.UpdateOutboxMessageAsSent(ctx, publishedMessage.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *BatchOutboxWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
