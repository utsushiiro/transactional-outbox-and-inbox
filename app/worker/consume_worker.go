package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type ConsumeWorker struct {
	db                ConsumeWorkerMessageDBDeps
	pollingInterval   time.Duration
	timeoutPerProcess time.Duration
	ticker            *timeutils.Ticker
	sleeper           *timeutils.RandomSleeper
}

type ConsumeWorkerMessageDBDeps struct {
	messagedb.Transactor
	inboxMessages messagedb.InboxMessages
}

func NewConsumeWorker(
	transactor messagedb.Transactor,
	inboxMessages messagedb.InboxMessages,
	pollingInterval time.Duration,
	timeoutPerProcess time.Duration,
) *ConsumeWorker {
	return &ConsumeWorker{
		db: ConsumeWorkerMessageDBDeps{
			Transactor:    transactor,
			inboxMessages: inboxMessages,
		},
		pollingInterval:   pollingInterval,
		timeoutPerProcess: timeoutPerProcess,
		sleeper:           timeutils.NewRandomSleeper(50*time.Millisecond, 200*time.Millisecond),
	}
}

func (c *ConsumeWorker) Run() error {
	ctx := context.Background()
	ticker := timeutils.NewTicker(c.pollingInterval)
	c.ticker = ticker

	for range ticker.C() {
		err := c.consumeMessage(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to consumeMessage", slog.String("error", err.Error()))
		}
	}

	return nil
}

func (c *ConsumeWorker) consumeMessage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeoutPerProcess)
	defer cancel()

	err := c.db.RunInTx(ctx, func(ctx context.Context) error {
		unprocessedMessage, err := c.db.inboxMessages.SelectUnprocessedOneWithLock(ctx)
		if err != nil {
			if errors.Is(err, messagedb.ErrResourceNotFound) {
				slog.InfoContext(ctx, "no unprocessed message")

				return nil
			}

			return err
		}

		// Perform some tasks in the same transaction with updating inbox message as processed.
		var msg string
		err = json.Unmarshal(unprocessedMessage.Payload, &msg)
		if err != nil {
			return err
		}
		c.sleeper.Sleep() // Simulate task processing time.
		slog.InfoContext(ctx, fmt.Sprintf("processed message: %v", msg))

		// After processing, mark the message as processed.
		unprocessedMessage.MarkAsProcessed()

		err = c.db.inboxMessages.Update(ctx, unprocessedMessage)
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
