package consumer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/message"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/msgclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/recovery"
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

	dbManager, err := rdb.NewDeprecatedSingleDBManager("postgres", "postgres", "localhost:5001", "transactional_outbox_and_inbox_example")
	if err != nil {
		log.Fatalf("failed to rdb.NewSingleDBManager: %v", err)
	}

	client, err := msgclient.NewSubscriber(mainCtx, "my-project", "my-subscription")
	if err != nil {
		log.Fatalf("failed to msgclient.NewSubscriber: %v", err)
	}

	inboxWorkerTimeoutPerProcess := 1 * time.Second
	inboxWorker := message.NewInboxWorker(dbManager, client, inboxWorkerTimeoutPerProcess)
	recovery.Go(inboxWorker.Run)

	consumeInterval := 100 * time.Millisecond
	consumeWorkerTimeoutPerProcess := 1 * time.Second
	consumeWorker := message.NewConsumeWorker(dbManager, consumeInterval, consumeWorkerTimeoutPerProcess)
	recovery.Go(consumeWorker.Run)

	ctx, cancel := signal.NotifyContext(mainCtx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	inboxWorker.Stop()
	consumeWorker.Stop()

	gracefulPeriod := max(inboxWorkerTimeoutPerProcess, consumeWorkerTimeoutPerProcess) + 1*time.Second
	time.Sleep(gracefulPeriod)

	log.Println("consumer stopped")
}
