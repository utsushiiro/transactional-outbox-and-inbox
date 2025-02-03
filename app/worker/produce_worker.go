package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/newmo-oss/ctxtime"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/domain/model"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/telemetry"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/timeutils"
)

type ProduceWorker struct {
	db                produceWorkerMessageDBDeps
	timeoutPerProcess time.Duration
	ticker            *timeutils.RandomTicker
}

type produceWorkerMessageDBDeps struct {
	messagedb.Transactor
	outboxMessages messagedb.OutboxMessages
}

func NewProduceWorker(
	transactor messagedb.Transactor,
	outboxMessages messagedb.OutboxMessages,
	timeoutPerProcess time.Duration,
) *ProduceWorker {
	return &ProduceWorker{
		db: produceWorkerMessageDBDeps{
			Transactor:     transactor,
			outboxMessages: outboxMessages,
		},
		timeoutPerProcess: timeoutPerProcess,
		ticker:            nil,
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
		// Start root span for each iteration.
		ctx, span := telemetry.Tracer.Start(ctx, "ProduceWorker.Run")

		err := p.produceMessage(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "failed to produceMessage", slog.String("error", err.Error()))
		}

		span.End()
	}

	return nil
}

func (p *ProduceWorker) produceMessage(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeoutPerProcess)
	defer cancel()

	err := p.db.RunInTx(ctx, func(ctx context.Context) error {
		// Perform some tasks here in the same transaction with inserting outbox message.

		// In a real-world scenario, this data would be related to the task.
		data, err := json.Marshal("Hello, World! at " + ctxtime.Now(ctx).In(time.Local).Format(time.RFC3339))
		if err != nil {
			return err
		}

		outboxMessage := model.NewOutboxMessage(data)

		err = p.db.outboxMessages.Insert(ctx, outboxMessage)
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

func (p *ProduceWorker) Stop() {
	if p.ticker != nil {
		p.ticker.Stop()
	}
}
