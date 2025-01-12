package worker

import (
	"context"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/domain/model"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/messagedb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/worker/mq"
)

type InboxWorker struct {
	db                inboxWorkerMessageDBDeps
	subscriber        mq.Subscriber
	timeoutPerProcess time.Duration
}

type inboxWorkerMessageDBDeps struct {
	messagedb.Transactor
	inboxMessages messagedb.InboxMessages
}

func NewInboxWorker(
	transactor messagedb.Transactor,
	inboxMessages messagedb.InboxMessages,
	subscriber mq.Subscriber,
	timeoutPerProcess time.Duration,
) *InboxWorker {
	return &InboxWorker{
		db: inboxWorkerMessageDBDeps{
			Transactor:    transactor,
			inboxMessages: inboxMessages,
		},
		subscriber:        subscriber,
		timeoutPerProcess: timeoutPerProcess,
	}
}

func (i *InboxWorker) Run() error {
	err := i.subscriber.Receive(context.Background(), func(ctx context.Context, msg *mq.Message, msgResponder mq.MessageResponder) {
		ctx, cancel := context.WithTimeout(ctx, i.timeoutPerProcess)
		defer cancel()

		err := i.db.RunInTx(ctx, func(ctx context.Context) error {
			inboxMessage := model.NewInboxMessage(msg.ID, msg.Payload)

			err := i.db.inboxMessages.Insert(ctx, inboxMessage)
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
