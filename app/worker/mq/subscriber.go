package mq

import (
	"context"
)

type Subscriber interface {
	Receive(ctx context.Context, handler func(ctx context.Context, message *Message) error) error
	Close() error
}

type AckableError struct {
	err error
}

func (a *AckableError) Error() string {
	return a.err.Error()
}

func (a *AckableError) Unwrap() error {
	return a.err
}

func MarkAsAckable(err error) error {
	if err == nil {
		return nil
	}

	return &AckableError{err: err}
}
