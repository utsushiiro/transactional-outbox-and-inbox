package pubsubclient

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
)

type subscriber struct {
	client       *pubsub.Client
	subscription *pubsub.Subscription
}

var _ mq.Subscriber = (*subscriber)(nil)

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

func (s *subscriber) Receive(ctx context.Context, handler func(context.Context, *mq.Message, mq.MessageResponder)) error {
	err := s.subscription.Receive(ctx, func(ctx context.Context, pubsubMsg *pubsub.Message) {
		msgID, err := uuid.Parse(pubsubMsg.Attributes["MessageID"])
		if err != nil {
			log.Printf("failed to uuid.Parse: %v", err)
			pubsubMsg.Ack()
			return
		}

		msg := &mq.Message{
			ID:      msgID,
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