package producer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/message"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/msgclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/recovery"
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

	dbManager, err := rdb.NewSingleDBManager("postgres", "postgres", "localhost:5001", "transactional_outbox_and_inbox_example")
	if err != nil {
		log.Fatalf("failed to rdb.NewSingleDBManager: %v", err)
	}

	client, err := msgclient.NewPublisher(mainCtx, "my-project", "my-topic")
	if err != nil {
		log.Fatalf("failed to publisher.New: %v", err)
	}

	outboxWorkerInterval := 1 * time.Second
	outboxWorker := message.NewOutboxWorker(dbManager, client, outboxWorkerInterval)
	recovery.Go(outboxWorker.Run)

	produceWorker := message.NewProduceWorker(dbManager)
	recovery.Go(produceWorker.Run)

	ctx, cancel := signal.NotifyContext(mainCtx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	produceWorker.Stop()
	time.Sleep(outboxWorkerInterval + 1*time.Second)
	outboxWorker.Stop()

	log.Println("producer stopped")
}
