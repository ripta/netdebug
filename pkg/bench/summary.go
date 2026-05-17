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

	Backends []BackendStats
}

// BackendStats summarises the requests sent to a single backend, identified
// by the fallback chain in backendKey. Source records which step in the
// chain produced the key so the report can render its provenance.
type BackendStats struct {
	Key            string
	Source         string
	Count          int
	ErrorCount     int
	PercentOfTotal float64
	P50            time.Duration
	P99            time.Duration
}
