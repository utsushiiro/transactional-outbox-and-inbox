package timeutils

import (
	"fmt"
	"math/rand"
	"time"
)

type RandomTicker struct {
	stop        chan struct{}
	c           chan time.Time
	randSrc     *rand.Rand
	minInterval time.Duration
	maxInterval time.Duration
}

func NewRandomTicker(minInterval time.Duration, maxInterval time.Duration) (*RandomTicker, error) {
	if minInterval > maxInterval {
		return nil, fmt.Errorf("maxInterval(=%d) should be greater than or equal to minInterval(=%d)", maxInterval, minInterval)
	}

	t := &RandomTicker{
		stop: make(chan struct{}),
		c:    make(chan time.Time),
		//nolint:gosec // We don't need a cryptographically secure random number generator here
		randSrc:     rand.New(rand.NewSource(NowUTC().UnixNano())),
		minInterval: minInterval,
		maxInterval: maxInterval,
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

func (r *RandomTicker) nextInterval() time.Duration {
	interval := r.randSrc.Int63n(int64(r.maxInterval-r.minInterval)) + int64(r.minInterval)

	return time.Duration(interval)
}

func (r *RandomTicker) C() <-chan time.Time {
	return r.c
}

func (r *RandomTicker) Stop() {
	close(r.stop)
}
