package mq

import (
	"context"

	"github.com/google/uuid"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/model"
)

type Publisher interface {
	Publish(ctx context.Context, msg *model.Message) error
	Close() error
}

type BatchPublisher interface {
	BatchPublish(ctx context.Context, msgs []*model.Message) (*BatchResult, error)
	Close() error
}

type BatchResult struct {
	SucceededIDs []uuid.UUID
	FailedIDs    []uuid.UUID
}
