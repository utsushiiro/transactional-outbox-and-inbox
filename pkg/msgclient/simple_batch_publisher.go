package msgclient

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/message"
)

type simpleBatchPublisher struct {
	client          *pubsub.Client
	topic           *pubsub.Topic
	workerLimitSize int
}

var _ message.BatchPublisher = (*simpleBatchPublisher)(nil)

func NewSimpleBatchPublisher(ctx context.Context, projectID string, topic string, workerLimitSize int) (*simpleBatchPublisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to pubsub.NewClient: %w", err)
	}

	return &simpleBatchPublisher{
		client:          client,
		topic:           client.Topic(topic),
		workerLimitSize: workerLimitSize,
	}, nil
}

func (p *simpleBatchPublisher) BatchPublish(ctx context.Context, messages []message.Message) ([]string, error) {
	// The `results` and `resultErrs` slices are shared across multiple goroutines,
	// but there is no race condition since each goroutine exclusively accesses its own index.
	results := make([]string, len(messages))
	resultErrs := make([]error, len(messages))

	wg := sync.WaitGroup{}
	limits := make(chan struct{}, p.workerLimitSize)

	for i, msg := range messages {
		pubsubMsg := &pubsub.Message{
			Attributes: map[string]string{
				"MessageID": msg.ID,
			},
			Data: msg.Payload,
		}

		wg.Add(1)
		limits <- struct{}{}

		go func() {
			defer func() {
				wg.Done()
				<-limits
			}()

			result := p.topic.Publish(ctx, pubsubMsg)
			_, err := result.Get(ctx)
			if err != nil {
				resultErrs[i] = fmt.Errorf("failed to publish message %s: %w", msg.ID, err)
			} else {
				results[i] = pubsubMsg.ID
			}
		}()
	}

	wg.Wait()

	var publishedIDs []string
	for _, result := range results {
		if result != "" {
			publishedIDs = append(publishedIDs, result)
		}
	}

	joinedErrors := errors.Join(resultErrs...)
	if joinedErrors != nil {
		return publishedIDs, joinedErrors
	}

	return publishedIDs, nil
}

func (p *simpleBatchPublisher) Close() error {
	return p.client.Close()
}
