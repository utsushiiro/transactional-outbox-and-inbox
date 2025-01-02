package msgclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type Publisher interface {
	Publish(ctx context.Context, msg string) (string, error)
	Close() error
}

type publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewPublisher(ctx context.Context, projectID string, topic string) (Publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to pubsub.NewClient: %w", err)
	}

	return &publisher{
		client: client,
		topic:  client.Topic(topic),
	}, nil
}

func (p *publisher) Publish(ctx context.Context, msg string) (string, error) {
	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	msgID, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to result.Get: %w", err)
	}

	return msgID, nil
}

func (p *publisher) Close() error {
	return p.client.Close()
}
