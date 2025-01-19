package pubsubclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/telemetry"
)

type publisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

var _ mq.Publisher = (*publisher)(nil)

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

func (p *publisher) Publish(ctx context.Context, message *mq.Message) error {
	ctx, span := telemetry.StartSpanWithFuncName(ctx)
	defer span.End()

	attributes := map[string]string{
		"MessageID": message.ID.String(),
	}

	// Inject otel context to message attributes.
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(attributes))

	result := p.topic.Publish(ctx, &pubsub.Message{
		Attributes: attributes,
		Data:       message.Payload,
	})

	_, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to result.Get: %w", err)
	}

	return nil
}

func (p *publisher) Close() error {
	return p.client.Close()
}
