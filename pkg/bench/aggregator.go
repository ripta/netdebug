package bench

import (
	"math"
	"sort"
	"time"
)

func aggregate(results []Result, elapsed time.Duration, connModel string) Summary {
	s := Summary{
		Count:     len(results),
		Elapsed:   elapsed,
		ConnModel: connModel,
	}

	successes := 0
	for _, r := range results {
		if r.Err != nil {
			s.ErrorCount++
			continue
		}
		successes++
	}

	if elapsed > 0 {
		s.Throughput = float64(s.Count) / elapsed.Seconds()
	}

	if successes == 0 {
		return s
	}

	total := make([]time.Duration, 0, successes)
	server := make([]time.Duration, 0, successes)
	network := make([]time.Duration, 0, successes)
	for _, r := range results {
		if r.Err != nil {
			continue
		}
		srv := time.Duration(r.ServerDurationNs)
		net := r.TotalDuration - srv
		if net < 0 {
			net = 0
		}
		total = append(total, r.TotalDuration)
		server = append(server, srv)
		network = append(network, net)
	}

	sort.Slice(total, func(i, j int) bool { return total[i] < total[j] })
	sort.Slice(server, func(i, j int) bool { return server[i] < server[j] })
	sort.Slice(network, func(i, j int) bool { return network[i] < network[j] })

	s.Total = computeLatencyStats(total)
	s.Server = computeLatencyStats(server)
	s.Network = computeLatencyStats(network)

	return s
}

func computeLatencyStats(sorted []time.Duration) LatencyStats {
	n := len(sorted)
	if n == 0 {
		return LatencyStats{}
	}

	var sum int64
	for _, d := range sorted {
		sum += int64(d)
	}
	mean := sum / int64(n)

	var sqDiff float64
	for _, d := range sorted {
		diff := float64(int64(d) - mean)
		sqDiff += diff * diff
	}
	stddev := time.Duration(math.Sqrt(sqDiff / float64(n)))

	return LatencyStats{
		Count:  n,
		Min:    sorted[0],
		Mean:   time.Duration(mean),
		Stddev: stddev,
		Max:    sorted[n-1],
		P50:    percentile(sorted, 50),
		P90:    percentile(sorted, 90),
		P99:    percentile(sorted, 99),
	}
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
