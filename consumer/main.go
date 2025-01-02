package main

import (
	"context"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/utsushiiro/transactional-outbox-and-inbox/consumer/pkg/msgclient"
)

func main() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:10002")

	client, err := msgclient.NewSubscriber(mainCtx, "my-project", "my-subscription")
	if err != nil {
		log.Fatalf("failed to msgclient.NewSubscriber: %v", err)
	}

	receiverCtx, cancel := context.WithTimeout(mainCtx, 10*time.Second)
	defer cancel()

	client.Receive(receiverCtx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		msg.Ack()
	})

	log.Println("Consumer stopped")
}
