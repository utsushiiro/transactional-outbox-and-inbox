package consumer

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/pubsubclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/recovery"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/telemetry"
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

	cleanup, err := telemetry.Setup(mainCtx, &telemetry.TelemetryConfig{
		ServiceName:        "github.com/utsushiiro/transactional-outbox-and-inbox/cmd/consumer",
		TracerAndMeterName: "consumer",
	})
	if err != nil {
		slog.ErrorContext(mainCtx, "failed to telemetry.Setup", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := cleanup(mainCtx); err != nil {
			slog.ErrorContext(mainCtx, "failed to cleanup", slog.String("error", err.Error()))
		}
	}()

	messageDB, err := messagedb.NewDB(mainCtx, "postgres", "postgres", "localhost:5001", "transactional_outbox_and_inbox_example")
	if err != nil {
		slog.ErrorContext(mainCtx, "failed to messagedb.NewDB", slog.String("error", err.Error()))
		os.Exit(1)
	}

	inboxMessages := messagedb.NewInboxMessages(messageDB)

	client, err := pubsubclient.NewSubscriber(mainCtx, "my-project", "my-subscription")
	if err != nil {
		slog.ErrorContext(mainCtx, "failed to pubsubclient.NewSubscriber", slog.String("error", err.Error()))
		os.Exit(1)
	}

	inboxWorkerTimeoutPerProcess := 1 * time.Second
	inboxWorker := worker.NewInboxWorker(messageDB, inboxMessages, client, inboxWorkerTimeoutPerProcess)
	recovery.Go(inboxWorker.Run)

	consumeInterval := 100 * time.Millisecond
	consumeWorkerTimeoutPerProcess := 1 * time.Second
	consumeWorker := worker.NewConsumeWorker(messageDB, inboxMessages, consumeInterval, consumeWorkerTimeoutPerProcess)
	recovery.Go(consumeWorker.Run)

	ctx, cancel := signal.NotifyContext(mainCtx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	inboxWorker.Stop()
	consumeWorker.Stop()

	gracefulPeriod := max(inboxWorkerTimeoutPerProcess, consumeWorkerTimeoutPerProcess) + 1*time.Second
	time.Sleep(gracefulPeriod)

	slog.InfoContext(ctx, "consumer stopped")
}
