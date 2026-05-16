package bench

import "time"

type Result struct {
	Start            time.Time
	End              time.Time
	TotalDuration    time.Duration
	ServerDurationNs int64
	Err              error
}
