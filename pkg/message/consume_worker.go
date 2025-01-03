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
	dbManager *rdb.SingleDBManager
	ticker    *timeutils.Ticker
}

func NewConsumeWorker(dbManager *rdb.SingleDBManager) *ConsumeWorker {
	return &ConsumeWorker{dbManager: dbManager}
}

func (c *ConsumeWorker) Run(ctx context.Context, interval time.Duration) {
	ticker := timeutils.NewTicker(interval)
	c.ticker = ticker

	for range ticker.C() {
		err := c.consumeMessage(ctx)
		if err != nil {
			log.Printf("failed to consumeMessage: %v", err)
		}
	}
}

func (c *ConsumeWorker) consumeMessage(ctx context.Context) error {
	err := c.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)

		unprocessedMessage, err := querier.SelectUnprocessedInboxMessage(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}

			return err
		}

		// process message
		var msg string
		err = json.Unmarshal(unprocessedMessage.MessagePayload, &msg)
		if err != nil {
			return err
		}
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
