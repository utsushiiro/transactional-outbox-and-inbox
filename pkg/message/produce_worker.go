package message

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
)

type ProduceWorker struct {
	dbManager         *rdb.SingleDBManager
	timeoutPerProcess time.Duration
	ticker            *timeutils.Ticker
}

func NewProduceWorker(dbManager *rdb.SingleDBManager, timeoutPerProcess time.Duration) *ProduceWorker {
	return &ProduceWorker{
		dbManager:         dbManager,
		timeoutPerProcess: timeoutPerProcess,
	}
}

func (p *ProduceWorker) Run() error {
	ctx := context.Background()
	ticker := timeutils.NewTicker(100 * time.Millisecond)
	p.ticker = ticker

	for range ticker.C() {
		err := p.produceMessage(ctx)
		if err != nil {
			log.Printf("failed to produceMessage: %v", err)
		}
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	}

	return nil
}

func (p *ProduceWorker) produceMessage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeoutPerProcess)
	defer cancel()

	data, err := json.Marshal("Hello, World! at " + time.Now().Format(time.RFC3339))
	if err != nil {
		return err
	}
	message := sqlc.InsertOutboxMessageParams{
		MessageTopic:   "my-topic",
		MessagePayload: data,
	}

	err = p.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)
		if _, err := querier.InsertOutboxMessage(ctx, message); err != nil {
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
