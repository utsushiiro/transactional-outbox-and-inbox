package pubsubclient

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/gammazero/workerpool"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
	"github.com/utsushiiro/transactional-outbox-and-inbox/pkg/telemetry"
)

type pooledBatchPublisher struct {
	client     *pubsub.Client
	topic      *pubsub.Topic
	workerPool *workerpool.WorkerPool
}

var _ mq.BatchPublisher = (*pooledBatchPublisher)(nil)

func NewPooledBatchPublisher(ctx context.Context, projectID string, topic string, workerPoolSize int) (*pooledBatchPublisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to pubsub.NewClient: %w", err)
	}

	return &pooledBatchPublisher{
		client:     client,
		topic:      client.Topic(topic),
		workerPool: workerpool.New(workerPoolSize),
	}, nil
}

func (p *pooledBatchPublisher) BatchPublish(ctx context.Context, messages []*mq.Message) (*mq.BatchResult, error) {
	ctx, span := telemetry.StartSpanWithFuncName(ctx)
	defer span.End()

	// The `errs` slice are shared across multiple goroutines,
	// but there is no race condition since each goroutine exclusively accesses its own index.
	errs := make([]error, len(messages))

	wg := sync.WaitGroup{}
	wg.Add(len(messages))

	for i, msg := range messages {
		p.workerPool.Submit(func() {
			defer wg.Done()

			pubsubMsg := &pubsub.Message{
				Attributes: map[string]string{
					"MessageID": msg.ID.String(),
				},
				Data: msg.Payload,
			}

			// Inject otel context to message attributes.
			otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(pubsubMsg.Attributes))

			result := p.topic.Publish(ctx, pubsubMsg)

			_, err := result.Get(ctx)
			if err != nil {
				errs[i] = fmt.Errorf("failed to publish message %s: %w", msg.ID, err)
			}
		})
	}

	wg.Wait()

	var (
		succeededIDs []uuid.UUID
		failedIDs    []uuid.UUID
	)
	for i, err := range errs {
		if err != nil {
			failedIDs = append(failedIDs, messages[i].ID)
		} else {
			succeededIDs = append(succeededIDs, messages[i].ID)
		}
	}
	batchResult := &mq.BatchResult{
		SucceededIDs: succeededIDs,
		FailedIDs:    failedIDs,
	}

	joinedErrors := errors.Join(errs...)
	if joinedErrors != nil {
		return batchResult, joinedErrors
	}

	return batchResult, nil
}

func (p *pooledBatchPublisher) Close() error {
	err := p.client.Close()
	if p.workerPool != nil {
		// NOTICE: If a lot of queued tasks are waiting, it may take a long time to stop.
		p.workerPool.StopWait()
	}
	// Since client.Close() returns an error interface, there is no risk of nil handling issues.
	return err
}
