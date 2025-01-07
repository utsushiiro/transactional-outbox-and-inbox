package message

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
)

type ProduceWorker struct {
	dbManager         *rdb.SingleDBManager
	timeoutPerProcess time.Duration
	ticker            *timeutils.RandomTicker
}

func NewProduceWorker(dbManager *rdb.SingleDBManager, timeoutPerProcess time.Duration) *ProduceWorker {
	return &ProduceWorker{
		dbManager:         dbManager,
		timeoutPerProcess: timeoutPerProcess,
	}
}

func (p *ProduceWorker) Run() error {
	ctx := context.Background()
	// Use a random ticker to simulate intervals between task occurrences.
	ticker, err := timeutils.NewRandomTicker(100*time.Millisecond, 500*time.Millisecond)
	if err != nil {
		return err
	}
	p.ticker = ticker

	for range ticker.C() {
		err := p.produceMessage(ctx)
		if err != nil {
			log.Printf("failed to produceMessage: %v", err)
		}
	}

	return nil
}

func (p *ProduceWorker) produceMessage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeoutPerProcess)
	defer cancel()

	err := p.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)

		// Perform some tasks here in the same transaction with inserting outbox message.

		// In a real-world scenario, this data would be related to the task.
		data, err := json.Marshal("Hello, World! at " + time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}

		if _, err := querier.InsertOutboxMessage(ctx, data); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *ProduceWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
