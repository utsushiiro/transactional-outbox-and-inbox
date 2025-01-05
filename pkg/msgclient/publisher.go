package msgclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/message"
)

type publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

var _ message.Publisher = (*publisher)(nil)

func NewPublisher(ctx context.Context, projectID string, topic string) (*publisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to pubsub.NewClient: %w", err)
	}

	return &publisher{
		client: client,
		topic:  client.Topic(topic),
	}, nil
}

func (p *publisher) Publish(ctx context.Context, msg message.Message) (string, error) {
	result := p.topic.Publish(ctx, &pubsub.Message{
		Attributes: map[string]string{
			"MessageID": msg.ID,
		},
		Data: msg.Payload,
	})

	pubsubMsgID, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to result.Get: %w", err)
	}

	return pubsubMsgID, nil
}

func (p *publisher) Close() error {
	return p.client.Close()
}
