package producer

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/msgclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

	dbManager, err := rdb.NewSingleDBManager("postgres", "postgres", "localhost:5001", "transactional_outbox_and_inbox_example")
	if err != nil {
		log.Fatalf("failed to NewSingleDBManager: %v", err)
	}

	data, err := json.Marshal("Hello, World!")
	if err != nil {
		log.Fatalf("failed to marshal: %v", err)
	}
	message := sqlc.InsertOutboxMessageParams{
		MessageTopic:   "my-topic",
		MessagePayload: data,
	}

	err = dbManager.RunInTx(mainCtx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)
		if _, err := querier.InsertOutboxMessage(ctx, message); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("failed to RunInTx: %v", err)
	}

	client, err := msgclient.NewPublisher(mainCtx, "my-project", "my-topic")
	if err != nil {
		log.Fatalf("failed to publisher.New: %v", err)
	}

	msgID, err := client.Publish(mainCtx, "Hello, World!")
	if err != nil {
		log.Fatalf("publish: %v", err)
	}

	log.Printf("Published a message; msg ID: %v\n", msgID)
}
