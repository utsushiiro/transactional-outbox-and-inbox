package timeutils

import (
	"math/rand"
	"time"
)

type RandomSleeper struct {
	randSrc     *rand.Rand
	minInterval time.Duration
	maxInterval time.Duration
}

func NewRandomSleeper(minInterval time.Duration, maxInterval time.Duration) *RandomSleeper {
	return &RandomSleeper{
		//nolint:gosec // We don't need a cryptographically secure random number generator here
		randSrc:     rand.New(rand.NewSource(time.Now().UnixNano())),
		minInterval: minInterval,
		maxInterval: maxInterval,
	}
}

func (rs *RandomSleeper) Sleep() {
	time.Sleep(rs.nextInterval())
}

func (rs *RandomSleeper) nextInterval() time.Duration {
	interval := rs.randSrc.Int63n(int64(rs.maxInterval-rs.minInterval)) + int64(rs.minInterval)

	return time.Duration(interval)
}
