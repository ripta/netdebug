package bench

import (
	"math"
	"sort"
	"time"
)

type Summary struct {
	Count      int
	ErrorCount int
	Elapsed    time.Duration
	Throughput float64
	LatencyP50 time.Duration
	LatencyP99 time.Duration
}

func summarize(results []Result, elapsed time.Duration) Summary {
	s := Summary{
		Count:   len(results),
		Elapsed: elapsed,
	}

	successes := make([]time.Duration, 0, len(results))
	for _, r := range results {
		if r.Err != nil {
			s.ErrorCount++
			continue
		}
		successes = append(successes, r.TotalDuration)
	}

	if elapsed > 0 {
		s.Throughput = float64(s.Count) / elapsed.Seconds()
	}

	if len(successes) > 0 {
		sort.Slice(successes, func(i, j int) bool { return successes[i] < successes[j] })
		s.LatencyP50 = percentile(successes, 50)
		s.LatencyP99 = percentile(successes, 99)
	}

	return s
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 100 {
		return sorted[len(sorted)-1]
	}
	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}
