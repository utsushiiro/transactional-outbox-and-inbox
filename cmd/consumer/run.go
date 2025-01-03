package consumer

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/msgclient"
)

func run() {
	mainCtx := context.Background()
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:5002")

	client, err := msgclient.NewSubscriber(mainCtx, "my-project", "my-subscription")
	if err != nil {
		log.Fatalf("failed to msgclient.NewSubscriber: %v", err)
	}

	receiverCtx, cancel := context.WithTimeout(mainCtx, 10*time.Second)
	defer cancel()

	client.Receive(receiverCtx, func(ctx context.Context, msg *msgclient.Message, msgResponder msgclient.MessageResponder) {
		log.Printf("Got message: id=%s payload=%s\n", msg.ID, string(msg.Payload))
		msgResponder.Ack()
	})

	log.Println("Consumer stopped")
}
