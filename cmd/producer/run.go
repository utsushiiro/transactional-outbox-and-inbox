package producer

import (
	"context"
	"log"
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
		log.Fatalf("failed to messagedb.NewDB: %v", err)
	}

	outboxMessages := messagedb.NewOutboxMessages(messageDB)

	client, err := pubsubclient.NewPublisher(mainCtx, "my-project", "my-topic")
	if err != nil {
		log.Fatalf("failed to pubsubclient.NewPublisher: %v", err)
	}

	batchClient, err := pubsubclient.NewPooledBatchPublisher(mainCtx, "my-project", "my-topic", 10)
	if err != nil {
		log.Fatalf("failed to pubsubclient.NewPooledBatchPublisher: %v", err)
	}

	outboxWorkerPoolingInterval := 1 * time.Second
	outboxWorkerTimeoutPerProcess := 1 * time.Second
	outboxWorker := worker.NewOutboxWorker(messageDB, outboxMessages, client, outboxWorkerPoolingInterval, outboxWorkerTimeoutPerProcess)
	recovery.Go(outboxWorker.Run)

	batchOutboxWorkerPoolingInterval := 1 * time.Second
	batchOutboxWorkerTimeoutPerProcess := 1 * time.Second
	batchOutboxWorker := worker.NewBatchOutboxWorker(messageDB, outboxMessages, batchClient, batchOutboxWorkerPoolingInterval, batchOutboxWorkerTimeoutPerProcess, 10)
	recovery.Go(batchOutboxWorker.Run)

	produceWorkerTimeoutPerProcess := 1 * time.Second
	produceWorker := worker.NewProduceWorker(messageDB, outboxMessages, produceWorkerTimeoutPerProcess)
	recovery.Go(produceWorker.Run)

	ctx, cancel := signal.NotifyContext(mainCtx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	produceWorker.Stop()
	outboxWorker.Stop()
	batchOutboxWorker.Stop()

	gracefulPeriod := max(outboxWorkerTimeoutPerProcess, produceWorkerTimeoutPerProcess) + 1*time.Second
	time.Sleep(gracefulPeriod)

	log.Println("producer stopped")
}
