package timeutils

import (
	"math/rand"
	"time"
)

type RandomSleeper struct {
	randSrc *rand.Rand
	min     time.Duration
	max     time.Duration
}

func NewRandomSleeper(min time.Duration, max time.Duration) *RandomSleeper {
	return &RandomSleeper{
		randSrc: rand.New(rand.NewSource(time.Now().UnixNano())),
		min:     min,
		max:     max,
	}
}

func (rs *RandomSleeper) Sleep() {
	time.Sleep(rs.nextInterval())
}

func (rs *RandomSleeper) nextInterval() time.Duration {
	interval := rs.randSrc.Int63n(int64(rs.max-rs.min)) + int64(rs.min)
	return time.Duration(interval)
}
