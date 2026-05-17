package bench

import (
	"math"
	"sort"
	"time"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// topErrorMessages is the maximum number of distinct error messages kept
// per status-code bucket. Three is enough to surface the dominant failure
// modes without flooding the report; an aggressive long tail still
// contributes to Count.
const topErrorMessages = 3

// maxErrorMessageLen is the rune length at which a message gets truncated
// for display. 80 fits a single terminal line after the code/count
// prefix; longer messages get a trailing "..." appended in place of the
// cut portion.
const maxErrorMessageLen = 80

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
	s.BackendSkew = computeBackendSkew(s.Backends)
	s.Errors = computeStatusCodes(results)

	return s
}

// computeStatusCodes groups errored results by gRPC status code, counts
// each distinct error message within the bucket, and returns up to the top
// topErrorMessages messages per code (truncated to maxErrorMessageLen
// runes). Non-status errors collapse into codes.Unknown so wrapped
// conn-acquire failures still get categorised. The returned slice is
// sorted by Count descending, then Code ascending, mirroring the
// per-backend ordering.
func computeStatusCodes(results []Result) []StatusCodeStats {
	if len(results) == 0 {
		return nil
	}

	type bucket struct {
		count    int
		messages map[string]int
	}
	buckets := make(map[codes.Code]*bucket)
	for _, r := range results {
		if r.Err == nil {
			continue
		}
		code := status.Code(r.Err)
		msg := errorMessage(r.Err)
		b, ok := buckets[code]
		if !ok {
			b = &bucket{messages: make(map[string]int)}
			buckets[code] = b
		}
		b.count++
		b.messages[truncateMessage(msg)]++
	}
	if len(buckets) == 0 {
		return nil
	}

	out := make([]StatusCodeStats, 0, len(buckets))
	for code, b := range buckets {
		out = append(out, StatusCodeStats{
			Code:        code,
			CodeName:    code.String(),
			Count:       b.count,
			TopMessages: topMessages(b.messages),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Code < out[j].Code
	})
	return out
}

// errorMessage extracts the human-readable message from an error. For gRPC
// status errors the Message() field carries the server-supplied string;
// for any other error type we fall back to Error() so wrapped failures
// from outside the call path (e.g., conn-acquire) still report something.
func errorMessage(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	return err.Error()
}

// truncateMessage clips a message at maxErrorMessageLen runes and appends
// "..." in place of the cut portion. Short messages pass through
// unchanged. Rune-counted rather than byte-counted so UTF-8 strings do
// not truncate mid-codepoint.
func truncateMessage(msg string) string {
	if utf8.RuneCountInString(msg) <= maxErrorMessageLen {
		return msg
	}
	runes := []rune(msg)
	return string(runes[:maxErrorMessageLen]) + "..."
}

// topMessages returns up to topErrorMessages entries from the message
// histogram, sorted by Count descending then Message ascending so output
// is deterministic when several messages tie on count.
func topMessages(messages map[string]int) []ErrorMessageStat {
	if len(messages) == 0 {
		return nil
	}
	stats := make([]ErrorMessageStat, 0, len(messages))
	for msg, c := range messages {
		stats = append(stats, ErrorMessageStat{Message: msg, Count: c})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count != stats[j].Count {
			return stats[i].Count > stats[j].Count
		}
		return stats[i].Message < stats[j].Message
	})
	if len(stats) > topErrorMessages {
		stats = stats[:topErrorMessages]
	}
	return stats
}

// computeBackendSkew measures imbalance across backends. CountRatio reports
// max/min request counts including error-only backends, since a backend
// that received only errors still received traffic. P99Ratio restricts to
// backends with at least one successful sample so a fully broken backend
// does not divide by zero. Either ratio is left at zero when fewer than
// two backends qualify.
func computeBackendSkew(backends []BackendStats) BackendSkew {
	var skew BackendSkew

	if len(backends) >= 2 {
		minC, maxC := backends[0].Count, backends[0].Count
		for _, b := range backends[1:] {
			if b.Count < minC {
				minC = b.Count
			}
			if b.Count > maxC {
				maxC = b.Count
			}
		}
		if minC > 0 {
			skew.CountRatio = float64(maxC) / float64(minC)
		}
	}

	var minP99, maxP99 time.Duration
	qualifying := 0
	for _, b := range backends {
		if b.P99 == 0 {
			continue
		}
		if qualifying == 0 || b.P99 < minP99 {
			minP99 = b.P99
		}
		if qualifying == 0 || b.P99 > maxP99 {
			maxP99 = b.P99
		}
		qualifying++
	}
	if qualifying >= 2 && minP99 > 0 {
		skew.P99Ratio = float64(maxP99) / float64(minP99)
	}

	return skew
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
