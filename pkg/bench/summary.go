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

	Backends    []BackendStats
	BackendSkew BackendSkew
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

// BackendSkew captures imbalance across backends. CountRatio is
// max-count/min-count across every backend that received traffic; one slow
// or one overloaded replica shows up as a large ratio. P99Ratio is
// max-p99/min-p99 across backends that recorded at least one successful
// sample, so a fully broken backend does not divide by zero. Either ratio
// is zero when fewer than two backends qualify; the report renders that as
// n/a.
type BackendSkew struct {
	CountRatio float64
	P99Ratio   float64
}
