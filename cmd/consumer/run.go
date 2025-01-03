package consumer

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
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

	dbManager, err := rdb.NewSingleDBManager("postgres", "postgres", "localhost:5001", "transactional_outbox_and_inbox_example")
	if err != nil {
		log.Fatalf("failed to rdb.NewSingleDBManager: %v", err)
	}

	client, err := msgclient.NewSubscriber(mainCtx, "my-project", "my-subscription")
	if err != nil {
		log.Fatalf("failed to msgclient.NewSubscriber: %v", err)
	}

	inboxWorker := message.NewInboxWorker(dbManager, client)
	go inboxWorker.Run(mainCtx)

	consumeWorker := message.NewConsumeWorker(dbManager)
	go consumeWorker.Run(mainCtx, 100*time.Millisecond)

	ctx, cancel := signal.NotifyContext(mainCtx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	inboxWorker.Stop()
	consumeWorker.Stop()

	log.Println("consumer stopped")
}
