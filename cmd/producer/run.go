package producer

import (
	"context"
	"log"
	"os"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/msgclient"
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:10002")

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
