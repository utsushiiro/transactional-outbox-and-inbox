package outbox

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/msgclient"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/rdb"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/sqlc"
	"github.com/utsushiiro/transactional-outbox-and-inbox/app/pkg/timeutils"
)

type Worker interface {
	Run(ctx context.Context, interval time.Duration)
	Stop()
}

type worker struct {
	dbManager *rdb.SingleDBManager
	publisher msgclient.Publisher
	ticker    *timeutils.Ticker
}

func NewWorker(dbManager *rdb.SingleDBManager, publisher msgclient.Publisher) Worker {
	return &worker{
		dbManager: dbManager,
		publisher: publisher,
	}
}

func (w *worker) Run(ctx context.Context, interval time.Duration) {
	ticker := timeutils.NewTicker(interval)
	w.ticker = ticker

	for range ticker.C() {
		w.run(ctx)
	}
}

func (w *worker) run(ctx context.Context) {
	_ = w.dbManager.RunInTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		querier := sqlc.NewQuerier(tx)

		unsentMessages, err := querier.SelectUnsentOutboxMessages(ctx, 10)
		if err != nil {
			return err
		}

		for _, unsentMessage := range unsentMessages {
			msgID, err := w.publisher.Publish(ctx, unsentMessage.MessagePayload)
			if err != nil {
				return err
			}

			log.Printf("Published message: %s", msgID)
		}

		for _, unsentMessage := range unsentMessages {
			updated, err := querier.UpdateOutboxMessageAsSent(ctx, unsentMessage.MessageUuid)
			if err != nil {
				return err
			}

			log.Printf("Updated messageUuid: %s", updated.MessageUuid)
		}

		return nil
	})
}

func (w *worker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
}
