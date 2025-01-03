package producer

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/msgclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/outbox"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
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

	worker := outbox.NewWorker(dbManager, client)
	go worker.Run(mainCtx, 1*time.Second)

	ticker := timeutils.NewTicker(150 * time.Millisecond)
	go insertMessagesToOutbox(mainCtx, dbManager, ticker)

	<-time.After(5 * time.Second)
	ticker.Stop()
	<-time.After(2 * time.Second)
	worker.Stop()
}

func insertMessagesToOutbox(ctx context.Context, dbManager *rdb.SingleDBManager, ticker *timeutils.Ticker) {
	for range ticker.C() {
		data, err := json.Marshal("Hello, World! at " + time.Now().Format(time.RFC3339))
		if err != nil {
			log.Fatalf("failed to marshal: %v", err)
		}
		message := sqlc.InsertOutboxMessageParams{
			MessageTopic:   "my-topic",
			MessagePayload: data,
		}

		err = dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
			querier := sqlc.NewQuerier(tx)
			if _, err := querier.InsertOutboxMessage(ctx, message); err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			log.Printf("failed to RunInTx: %v", err)
		}
	}
}
