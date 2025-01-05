package message

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
)

type InboxWorker struct {
	dbManager         *rdb.SingleDBManager
	subscriber        Subscriber
	timeoutPerProcess time.Duration
}

func NewInboxWorker(
	dbManager *rdb.SingleDBManager,
	subscriber Subscriber,
	timeoutPerProcess time.Duration,
) *InboxWorker {
	return &InboxWorker{
		dbManager:         dbManager,
		subscriber:        subscriber,
		timeoutPerProcess: timeoutPerProcess,
	}
}

func (i *InboxWorker) Run() error {
	err := i.subscriber.Receive(context.Background(), func(ctx context.Context, msg *Message, msgResponder MessageResponder) {
		ctx, cancel := context.WithTimeout(ctx, i.timeoutPerProcess)
		defer cancel()

		err := i.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
			querier := sqlc.NewQuerier(tx)

			parsedMsgID, err := uuid.Parse(msg.ID)
			if err != nil {
				return err
			}

			_, err = querier.InsertInboxMessage(ctx, sqlc.InsertInboxMessageParams{
				MessageUuid:    parsedMsgID,
				MessagePayload: msg.Payload,
			})
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			log.Printf("failed to insert inbox message: %v", err)
			msgResponder.Nack()
			return
		}

		msgResponder.Ack()
	})

	return err
}

func (i *InboxWorker) Stop() {
	i.subscriber.Close()
}
