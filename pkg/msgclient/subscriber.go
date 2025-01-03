package msgclient

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

type Subscriber interface {
	Receive(ctx context.Context, handler func(context.Context, *Message, MessageResponder)) error
	Close() error
}

type subscriber struct {
	client       *pubsub.Client
	subscription *pubsub.Subscription
}

func NewSubscriber(
	ctx context.Context,
	projectID string,
	subscriptionID string,
) (*subscriber, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to pubsub.NewClient: %w", err)
	}

	return &subscriber{
		client:       client,
		subscription: client.Subscription(subscriptionID),
	}, nil
}

func (s *subscriber) Receive(ctx context.Context, handler func(context.Context, *Message, MessageResponder)) error {
	err := s.subscription.Receive(ctx, func(ctx context.Context, pubsubMsg *pubsub.Message) {
		msg := &Message{
			ID:      pubsubMsg.Attributes["MessageID"],
			Payload: pubsubMsg.Data,
		}
		handler(ctx, msg, pubsubMsg)
	})
	if err != nil {
		return fmt.Errorf("failed to subscription.Receive: %w", err)
	}

	return nil
}

func (s *subscriber) Close() error {
	return s.client.Close()
}
