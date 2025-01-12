package worker

import (
	"context"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/infra/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/model"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
)

type InboxWorker struct {
	db                *messagedb.DB
	subscriber        mq.Subscriber
	timeoutPerProcess time.Duration
}

func NewInboxWorker(
	db *messagedb.DB,
	subscriber mq.Subscriber,
	timeoutPerProcess time.Duration,
) *InboxWorker {
	return &InboxWorker{
		db:                db,
		subscriber:        subscriber,
		timeoutPerProcess: timeoutPerProcess,
	}
}

func (i *InboxWorker) Run() error {
	err := i.subscriber.Receive(context.Background(), func(ctx context.Context, msg *model.Message, msgResponder mq.MessageResponder) {
		ctx, cancel := context.WithTimeout(ctx, i.timeoutPerProcess)
		defer cancel()

		err := i.db.RunInTx(ctx, func(ctx context.Context) error {
			err := i.db.InsertInboxMessage(ctx, &messagedb.InsertInboxMessageParams{
				MessageID: msg.ID,
				Payload:   msg.Payload,
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
