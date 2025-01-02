package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
)

func main() {
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:10002")
	err := publish("my-project", "my-topic", "Hello World")
	if err != nil {
		log.Fatalf("publish: %v", err)
	}
}

func publish(projectID, topicID, msg string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub: NewClient: %w", err)
	}
	defer client.Close()

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %w", err)
	}

	log.Printf("Published a message; msg ID: %v\n", id)
	return nil
}
