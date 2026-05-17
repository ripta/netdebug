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

	if successes > 0 {
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
	}

	s.Backends = computeBackends(results, s.Count)

	return s
}

// computeBackends groups results by backendKey and emits one BackendStats
// per group. Errors count toward the group's Count and ErrorCount but their
// durations do not contribute to the p50/p99 sample; an all-errors group
// reports zero percentiles. The slice is sorted by Count descending and
// then Key ascending so output ordering is deterministic across runs.
func computeBackends(results []Result, totalCount int) []BackendStats {
	if len(results) == 0 {
		return nil
	}

	type bucket struct {
		Source     string
		Count      int
		ErrorCount int
		Durations  []time.Duration
	}
	buckets := make(map[string]*bucket)
	for _, r := range results {
		key, src := backendKey(r)
		b, ok := buckets[key]
		if !ok {
			b = &bucket{Source: src}
			buckets[key] = b
		}
		b.Count++
		if r.Err != nil {
			b.ErrorCount++
			continue
		}
		b.Durations = append(b.Durations, r.TotalDuration)
	}

	out := make([]BackendStats, 0, len(buckets))
	for key, b := range buckets {
		sort.Slice(b.Durations, func(i, j int) bool { return b.Durations[i] < b.Durations[j] })
		stat := BackendStats{
			Key:        key,
			Source:     b.Source,
			Count:      b.Count,
			ErrorCount: b.ErrorCount,
			P50:        percentile(b.Durations, 50),
			P99:        percentile(b.Durations, 99),
		}
		if totalCount > 0 {
			stat.PercentOfTotal = float64(b.Count) / float64(totalCount) * 100
		}
		out = append(out, stat)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Key < out[j].Key
	})
	return out
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

// backendKey picks the per-backend grouping key for a request, walking the
// fallback chain settled in NDB-004: kubernetes pod_name, then kubernetes
// hostname, then the resolved peer address. The second return identifies
// which step in the chain produced the key so the report can render the
// provenance alongside it. Results with no identifier at all share an
// "unknown" bucket; this is the right behaviour under bufconn tests and
// when an RPC fails before any peer is resolved.
func backendKey(r Result) (key, source string) {
	switch {
	case r.PodName != "":
		return r.PodName, "pod_name"
	case r.PodHostname != "":
		return r.PodHostname, "hostname"
	case r.PeerAddr != "":
		return r.PeerAddr, "peer"
	default:
		return "", "unknown"
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
