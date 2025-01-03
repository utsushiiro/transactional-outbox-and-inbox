package outbox

import (
	"time"
)

type ticker struct {
	base *time.Ticker
	stop chan struct{}
	c    chan time.Time
}

func NewTicker(d time.Duration) *ticker {
	t := &ticker{
		base: time.NewTicker(d),
		stop: make(chan struct{}),
		c:    make(chan time.Time),
	}

	go func() {
		for {
			select {
			case v := <-t.base.C:
				t.c <- v
			case <-t.stop:
				close(t.c)
				return
			}
		}
	}()

	return t
}

func (t *ticker) C() <-chan time.Time {
	return t.c
}

func (t *ticker) Stop() {
	t.base.Stop()
	close(t.stop)
}
