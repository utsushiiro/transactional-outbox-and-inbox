package timeutils

import "time"

func NowUTC() time.Time {
	//nolint:forbidigo
	return time.Now().UTC()
}
