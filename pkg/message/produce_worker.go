package message

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type ProduceWorker struct {
	db                *messagedb.DB
	timeoutPerProcess time.Duration
	ticker            *timeutils.RandomTicker
}

func NewProduceWorker(db *messagedb.DB, timeoutPerProcess time.Duration) *ProduceWorker {
	return &ProduceWorker{
		db:                db,
		timeoutPerProcess: timeoutPerProcess,
	}
}

func (p *ProduceWorker) Run() error {
	ctx := context.Background()
	// Use a random ticker to simulate intervals between task occurrences.
	ticker, err := timeutils.NewRandomTicker(50*time.Millisecond, 200*time.Millisecond)
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

	err := p.db.RunInTx(ctx, func(ctx context.Context) error {
		// Perform some tasks here in the same transaction with inserting outbox message.

		// In a real-world scenario, this data would be related to the task.
		data, err := json.Marshal("Hello, World! at " + time.Now().Format(time.RFC3339))
		if err != nil {
			return err
		}

		if err := p.db.InsertOutboxMessage(ctx, data); err != nil {
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
