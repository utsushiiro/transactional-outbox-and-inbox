package producer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/message"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/msgclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/recovery"
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

	messageDB, err := messagedb.NewDB(mainCtx, "postgres", "postgres", "localhost:5001", "transactional_outbox_and_inbox_example")
	if err != nil {
		log.Fatalf("failed to messagedb.NewDB: %v", err)
	}

	client, err := msgclient.NewPublisher(mainCtx, "my-project", "my-topic")
	if err != nil {
		log.Fatalf("failed to msgclient.NewPublisher: %v", err)
	}

	batchClient, err := msgclient.NewPooledBatchPublisher(mainCtx, "my-project", "my-topic", 10)
	if err != nil {
		log.Fatalf("failed to msgclient.NewPooledBatchPublisher: %v", err)
	}

	outboxWorkerPoolingInterval := 1 * time.Second
	outboxWorkerTimeoutPerProcess := 1 * time.Second
	outboxWorker := message.NewOutboxWorker(messageDB, client, outboxWorkerPoolingInterval, outboxWorkerTimeoutPerProcess)
	recovery.Go(outboxWorker.Run)

	batchOutboxWorkerPoolingInterval := 1 * time.Second
	batchOutboxWorkerTimeoutPerProcess := 1 * time.Second
	batchOutboxWorker := message.NewBatchOutboxWorker(messageDB, batchClient, batchOutboxWorkerPoolingInterval, batchOutboxWorkerTimeoutPerProcess, 10)
	recovery.Go(batchOutboxWorker.Run)

	produceWorkerTimeoutPerProcess := 1 * time.Second
	produceWorker := message.NewProduceWorker(messageDB, produceWorkerTimeoutPerProcess)
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
