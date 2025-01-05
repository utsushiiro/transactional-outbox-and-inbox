package timeutils

import (
	"fmt"
	"math/rand"
	"time"
)

type RandomTicker struct {
	stop    chan struct{}
	c       chan time.Time
	randSrc *rand.Rand
	min     time.Duration
	max     time.Duration
}

func NewRandomTicker(min time.Duration, max time.Duration) (*RandomTicker, error) {
	if min > max {
		return nil, fmt.Errorf("max(=%d) should be greater than or equal to min(=%d)", max, min)
	}

	t := &RandomTicker{
		stop:    make(chan struct{}),
		c:       make(chan time.Time),
		randSrc: rand.New(rand.NewSource(time.Now().UnixNano())),
		min:     min,
		max:     max,
	}

	go func() {
		for {
			select {
			case v := <-time.After(t.nextInterval()):
				t.c <- v
			case <-t.stop:
				close(t.c)
				return
			}
		}
	}()

	return t, nil
}

func (rt *RandomTicker) nextInterval() time.Duration {
	interval := rt.randSrc.Int63n(int64(rt.max-rt.min)) + int64(rt.min)
	return time.Duration(interval)
}

func (t *RandomTicker) C() <-chan time.Time {
	return t.c
}

func (t *RandomTicker) Stop() {
	close(t.stop)
}
