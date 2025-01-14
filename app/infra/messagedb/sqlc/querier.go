// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package sqlc

import (
	"context"
)

type Querier interface {
	BulkUpdateOutboxMessagesAsSent(ctx context.Context, arg BulkUpdateOutboxMessagesAsSentParams) error
	InsertInboxMessage(ctx context.Context, arg InsertInboxMessageParams) error
	InsertOutboxMessage(ctx context.Context, arg InsertOutboxMessageParams) error
	SelectUnprocessedInboxMessage(ctx context.Context) (InboxMessage, error)
	SelectUnsentOutboxMessage(ctx context.Context) (OutboxMessage, error)
	SelectUnsentOutboxMessages(ctx context.Context, limit int32) ([]OutboxMessage, error)
	UpdateInboxMessage(ctx context.Context, arg UpdateInboxMessageParams) error
	UpdateOutboxMessage(ctx context.Context, arg UpdateOutboxMessageParams) error
}

var _ Querier = (*Queries)(nil)