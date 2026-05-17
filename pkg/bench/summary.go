package bench

import "time"

type LatencyStats struct {
	Count  int
	Min    time.Duration
	Mean   time.Duration
	Stddev time.Duration
	Max    time.Duration
	P50    time.Duration
	P90    time.Duration
	P99    time.Duration
}

type Summary struct {
	Count      int
	ErrorCount int
	Elapsed    time.Duration
	Throughput float64
	ConnModel  string

	Total   LatencyStats
	Server  LatencyStats
	Network LatencyStats
}
