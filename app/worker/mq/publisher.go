package mq

import (
	"context"

	"github.com/google/uuid"
)

type Publisher interface {
	Publish(ctx context.Context, msg *Message) error
	Close() error
}

type BatchPublisher interface {
	BatchPublish(ctx context.Context, msgs []*Message) (*BatchResult, error)
	Close() error
}

type BatchResult struct {
	SucceededIDs []uuid.UUID
	FailedIDs    []uuid.UUID
}
