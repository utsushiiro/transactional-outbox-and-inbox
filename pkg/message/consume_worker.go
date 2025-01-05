package message

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
)

type ConsumeWorker struct {
	dbManager         *rdb.SingleDBManager
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	ticker            *timeutils.Ticker
	sleeper           *timeutils.RandomSleeper
}

func NewConsumeWorker(
	dbManager *rdb.SingleDBManager,
	pollingInterval time.Duration,
	timeoutPerProcess time.Duration,
) *ConsumeWorker {
	return &ConsumeWorker{
		dbManager:         dbManager,
		pollingInterval:   pollingInterval,
		timeoutPerProcess: timeoutPerProcess,
		sleeper:           timeutils.NewRandomSleeper(100*time.Millisecond, 500*time.Millisecond),
	}
}

func (c *ConsumeWorker) Run() error {
	ctx := context.Background()
	ticker := timeutils.NewTicker(c.pollingInterval)
	c.ticker = ticker

	for range ticker.C() {
		err := c.consumeMessage(ctx)
		if err != nil {
			log.Printf("failed to consumeMessage: %v", err)
		}
	}

	return nil
}

func (c *ConsumeWorker) consumeMessage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeoutPerProcess)
	defer cancel()

	err := c.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)

		unprocessedMessage, err := querier.SelectUnprocessedInboxMessage(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			return err
		}

		// do some tasks in the same transaction with updating inbox message as processed
		var msg string
		err = json.Unmarshal(unprocessedMessage.MessagePayload, &msg)
		if err != nil {
			return err
		}
		c.sleeper.Sleep() // simulate task processing time
		log.Printf("processed message: %v", msg)

		_, err = querier.UpdateInboxMessageAsProcessed(ctx, unprocessedMessage.MessageUuid)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsumeWorker) Stop() {
	c.ticker.Stop()
}
