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
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

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
