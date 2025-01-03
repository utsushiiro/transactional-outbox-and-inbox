package timeutils

import (
	"time"
)

type Ticker struct {
	base *time.Ticker
	stop chan struct{}
	c    chan time.Time
}

func NewTicker(d time.Duration) *Ticker {
	t := &Ticker{
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

func (t *Ticker) C() <-chan time.Time {
	return t.c
}

func (t *Ticker) Stop() {
	t.base.Stop()
	close(t.stop)
}
