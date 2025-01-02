package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
)

func main() {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:10002")
	err := pullMsgs("my-project", "my-subscription")
	if err != nil {
		log.Fatalf("pullMsgs: %v", err)
	}
}

func pullMsgs(projectID, subID string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %w", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))
		atomic.AddInt32(&received, 1)
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %w", err)
	}
	log.Printf("Received %d messages\n", received)

	return nil
}
