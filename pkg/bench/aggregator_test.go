package bench

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

type computeStatsTest struct {
	Name   string
	Sorted []time.Duration
	Want   LatencyStats
}

var computeStatsTests = []computeStatsTest{
	{
		Name:   "empty input returns zero value",
		Sorted: nil,
		Want:   LatencyStats{},
	},
	{
		Name:   "single sample mirrors itself across every metric",
		Sorted: []time.Duration{ms(7)},
		Want: LatencyStats{
			Count: 1, Min: ms(7), Mean: ms(7), Stddev: 0, Max: ms(7),
			P50: ms(7), P90: ms(7), P99: ms(7),
		},
	},
	{
		// Population variance of {1,2,3,4,5} = 2; stddev = sqrt(2).
		Name:   "five-sample distribution has hand-calculable stddev",
		Sorted: []time.Duration{ms(1), ms(2), ms(3), ms(4), ms(5)},
		Want: LatencyStats{
			Count:  5,
			Min:    ms(1),
			Mean:   ms(3),
			Stddev: time.Duration(math.Sqrt(2e12)),
			Max:    ms(5),
			P50:    ms(3),
			P90:    ms(5),
			P99:    ms(5),
		},
	},
}

func TestComputeLatencyStats(t *testing.T) {
	for _, tc := range computeStatsTests {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Want, computeLatencyStats(tc.Sorted))
		})
	}
}

func TestComputeLatencyStats_HundredAscending(t *testing.T) {
	got := computeLatencyStats(ascendingDurations(100))
	assert.Equal(t, 100, got.Count)
	assert.Equal(t, ms(1), got.Min)
	assert.Equal(t, ms(100), got.Max)
	assert.Equal(t, ms(50)+500*time.Microsecond, got.Mean)
	assert.Equal(t, ms(50), got.P50)
	assert.Equal(t, ms(90), got.P90)
	assert.Equal(t, ms(99), got.P99)
	// Population variance of integers 1..100 = (100^2 - 1) / 12 = 833.25.
	// Scaled to ns^2 the variance is 833.25e12.
	assert.InDelta(t, math.Sqrt(833.25e12), float64(got.Stddev), 1)
}

type aggregateTest struct {
	Name             string
	Results          []Result
	Elapsed          time.Duration
	WantCount        int
	WantErrors       int
	WantThrPS        float64
	WantTotalCount   int
	WantTotalMin     time.Duration
	WantTotalMax     time.Duration
	WantTotalP50     time.Duration
	WantServerCount  int
	WantServerMin    time.Duration
	WantServerMax    time.Duration
	WantNetworkCount int
	WantNetworkMin   time.Duration
	WantNetworkMax   time.Duration
}

var aggregateTests = []aggregateTest{
	{
		Name:    "empty input produces zero summary",
		Results: nil,
		Elapsed: time.Second,
	},
	{
		Name: "errors-only run leaves every latency series empty",
		Results: []Result{
			{Err: errors.New("boom")},
			{Err: errors.New("boom")},
			{Err: errors.New("boom")},
		},
		Elapsed:    time.Second,
		WantCount:  3,
		WantErrors: 3,
		WantThrPS:  3,
	},
	{
		Name: "successes split into total, server, network independently",
		Results: []Result{
			{TotalDuration: ms(5), ServerDurationNs: int64(ms(1))},
			{TotalDuration: ms(7), ServerDurationNs: int64(ms(2))},
			{TotalDuration: ms(9), ServerDurationNs: int64(ms(3))},
		},
		Elapsed:          time.Second,
		WantCount:        3,
		WantThrPS:        3,
		WantTotalCount:   3,
		WantTotalMin:     ms(5),
		WantTotalMax:     ms(9),
		WantTotalP50:     ms(7),
		WantServerCount:  3,
		WantServerMin:    ms(1),
		WantServerMax:    ms(3),
		WantNetworkCount: 3,
		WantNetworkMin:   ms(4),
		WantNetworkMax:   ms(6),
	},
	{
		Name: "errors skip latency series, throughput counts everything",
		Results: []Result{
			{TotalDuration: ms(1)},
			{Err: errors.New("boom")},
			{TotalDuration: ms(3)},
			{Err: errors.New("boom")},
			{TotalDuration: ms(5)},
		},
		Elapsed:          time.Second,
		WantCount:        5,
		WantErrors:       2,
		WantThrPS:        5,
		WantTotalCount:   3,
		WantTotalMin:     ms(1),
		WantTotalMax:     ms(5),
		WantTotalP50:     ms(3),
		WantServerCount:  3,
		WantNetworkCount: 3,
		WantNetworkMin:   ms(1),
		WantNetworkMax:   ms(5),
	},
	{
		// When the server-reported handler time exceeds the client-observed
		// RTT, network time clamps to zero rather than reporting a negative
		// duration. This is a sub-microsecond artifact in bufconn tests.
		Name: "network clamps to zero when server exceeds total",
		Results: []Result{
			{TotalDuration: ms(5), ServerDurationNs: int64(ms(10))},
		},
		Elapsed:          time.Second,
		WantCount:        1,
		WantThrPS:        1,
		WantTotalCount:   1,
		WantTotalMin:     ms(5),
		WantTotalMax:     ms(5),
		WantTotalP50:     ms(5),
		WantServerCount:  1,
		WantServerMin:    ms(10),
		WantServerMax:    ms(10),
		WantNetworkCount: 1,
		WantNetworkMin:   0,
		WantNetworkMax:   0,
	},
	{
		Name:             "zero elapsed produces zero throughput",
		Results:          []Result{{TotalDuration: ms(1)}},
		Elapsed:          0,
		WantCount:        1,
		WantThrPS:        0,
		WantTotalCount:   1,
		WantTotalMin:     ms(1),
		WantTotalMax:     ms(1),
		WantTotalP50:     ms(1),
		WantServerCount:  1,
		WantNetworkCount: 1,
		WantNetworkMin:   ms(1),
		WantNetworkMax:   ms(1),
	},
	{
		Name:             "100 ascending samples pick nearest-rank percentiles",
		Results:          ascendingResults(100),
		Elapsed:          time.Second,
		WantCount:        100,
		WantThrPS:        100,
		WantTotalCount:   100,
		WantTotalMin:     ms(1),
		WantTotalMax:     ms(100),
		WantTotalP50:     ms(50),
		WantServerCount:  100,
		WantNetworkCount: 100,
		WantNetworkMin:   ms(1),
		WantNetworkMax:   ms(100),
	},
}

func TestAggregate(t *testing.T) {
	for _, tc := range aggregateTests {
		t.Run(tc.Name, func(t *testing.T) {
			s := aggregate(tc.Results, tc.Elapsed, ConnModelPerWorker)
			assert.Equal(t, tc.WantCount, s.Count, "Count")
			assert.Equal(t, tc.WantErrors, s.ErrorCount, "ErrorCount")
			assert.InDelta(t, tc.WantThrPS, s.Throughput, 1e-9, "Throughput")
			assert.Equal(t, tc.Elapsed, s.Elapsed, "Elapsed")
			assert.Equal(t, ConnModelPerWorker, s.ConnModel, "ConnModel")

			assert.Equal(t, tc.WantTotalCount, s.Total.Count, "Total.Count")
			assert.Equal(t, tc.WantTotalMin, s.Total.Min, "Total.Min")
			assert.Equal(t, tc.WantTotalMax, s.Total.Max, "Total.Max")
			assert.Equal(t, tc.WantTotalP50, s.Total.P50, "Total.P50")

			assert.Equal(t, tc.WantServerCount, s.Server.Count, "Server.Count")
			assert.Equal(t, tc.WantServerMin, s.Server.Min, "Server.Min")
			assert.Equal(t, tc.WantServerMax, s.Server.Max, "Server.Max")

			assert.Equal(t, tc.WantNetworkCount, s.Network.Count, "Network.Count")
			assert.Equal(t, tc.WantNetworkMin, s.Network.Min, "Network.Min")
			assert.Equal(t, tc.WantNetworkMax, s.Network.Max, "Network.Max")
		})
	}
}

type percentileTest struct {
	Name   string
	Sorted []time.Duration
	P      float64
	Want   time.Duration
}

var percentileTests = []percentileTest{
	{Name: "empty returns zero", Sorted: nil, P: 50, Want: 0},
	{Name: "p<=0 returns first", Sorted: []time.Duration{1, 2, 3}, P: 0, Want: 1},
	{Name: "p>=100 returns last", Sorted: []time.Duration{1, 2, 3}, P: 100, Want: 3},
	{Name: "p50 of 100 is rank 50", Sorted: ascendingDurations(100), P: 50, Want: ms(50)},
	{Name: "p99 of 100 is rank 99", Sorted: ascendingDurations(100), P: 99, Want: ms(99)},
}

func TestPercentile(t *testing.T) {
	for _, tc := range percentileTests {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.Want, percentile(tc.Sorted, tc.P))
		})
	}
}

type backendBucket struct {
	Key            string
	Source         string
	Count          int
	ErrorCount     int
	PercentOfTotal float64
	P50            time.Duration
	P99            time.Duration
}

type aggregateBackendsTest struct {
	Name    string
	Results []Result
	Want    []backendBucket
}

var aggregateBackendsTests = []aggregateBackendsTest{
	{
		Name:    "empty input produces no backends",
		Results: nil,
		Want:    nil,
	},
	{
		Name: "pod_name groups together",
		Results: []Result{
			{TotalDuration: ms(2), PodName: "pod-a"},
			{TotalDuration: ms(4), PodName: "pod-a"},
			{TotalDuration: ms(6), PodName: "pod-a"},
		},
		Want: []backendBucket{
			{Key: "pod-a", Source: "pod_name", Count: 3, PercentOfTotal: 100, P50: ms(4), P99: ms(6)},
		},
	},
	{
		Name: "hostname fallback when pod_name absent",
		Results: []Result{
			{TotalDuration: ms(3), PodHostname: "host-b", PeerAddr: "ignored"},
			{TotalDuration: ms(5), PodHostname: "host-b", PeerAddr: "ignored"},
		},
		Want: []backendBucket{
			{Key: "host-b", Source: "hostname", Count: 2, PercentOfTotal: 100, P50: ms(3), P99: ms(5)},
		},
	},
	{
		Name: "peer fallback when both kubernetes fields absent",
		Results: []Result{
			{TotalDuration: ms(7), PeerAddr: "10.0.0.1:50051"},
		},
		Want: []backendBucket{
			{Key: "10.0.0.1:50051", Source: "peer", Count: 1, PercentOfTotal: 100, P50: ms(7), P99: ms(7)},
		},
	},
	{
		// Each result carries a different fallback source; grouping must
		// not collapse them by accident even though every entry holds
		// "10.0.0.1:50051" as PeerAddr.
		Name: "mixed fallback sources group separately",
		Results: []Result{
			{TotalDuration: ms(1), PodName: "pod-a", PeerAddr: "10.0.0.1:50051"},
			{TotalDuration: ms(2), PodHostname: "host-b", PeerAddr: "10.0.0.1:50051"},
			{TotalDuration: ms(3), PeerAddr: "10.0.0.1:50051"},
		},
		Want: []backendBucket{
			{Key: "10.0.0.1:50051", Source: "peer", Count: 1, PercentOfTotal: 100.0 / 3.0, P50: ms(3), P99: ms(3)},
			{Key: "host-b", Source: "hostname", Count: 1, PercentOfTotal: 100.0 / 3.0, P50: ms(2), P99: ms(2)},
			{Key: "pod-a", Source: "pod_name", Count: 1, PercentOfTotal: 100.0 / 3.0, P50: ms(1), P99: ms(1)},
		},
	},
	{
		Name: "errors counted per backend without dragging percentiles",
		Results: []Result{
			{TotalDuration: ms(4), PodName: "pod-a"},
			{Err: errors.New("boom"), PodName: "pod-a"},
			{TotalDuration: ms(8), PodName: "pod-b"},
			{Err: errors.New("boom"), PodName: "pod-b"},
			{Err: errors.New("boom"), PodName: "pod-b"},
		},
		Want: []backendBucket{
			{Key: "pod-b", Source: "pod_name", Count: 3, ErrorCount: 2, PercentOfTotal: 60, P50: ms(8), P99: ms(8)},
			{Key: "pod-a", Source: "pod_name", Count: 2, ErrorCount: 1, PercentOfTotal: 40, P50: ms(4), P99: ms(4)},
		},
	},
	{
		Name: "errors-only group reports zero percentiles",
		Results: []Result{
			{Err: errors.New("boom"), PeerAddr: "10.0.0.1:50051"},
			{Err: errors.New("boom"), PeerAddr: "10.0.0.1:50051"},
		},
		Want: []backendBucket{
			{Key: "10.0.0.1:50051", Source: "peer", Count: 2, ErrorCount: 2, PercentOfTotal: 100, P50: 0, P99: 0},
		},
	},
	{
		Name: "results with no identifier collapse into unknown",
		Results: []Result{
			{TotalDuration: ms(1)},
			{TotalDuration: ms(3)},
		},
		Want: []backendBucket{
			{Key: "", Source: "unknown", Count: 2, PercentOfTotal: 100, P50: ms(1), P99: ms(3)},
		},
	},
	{
		// Sort tie-breaker: identical counts sort by Key asc. pod-a and
		// pod-b both carry one request; pod-a must come first.
		Name: "tied counts sort by key ascending",
		Results: []Result{
			{TotalDuration: ms(2), PodName: "pod-b"},
			{TotalDuration: ms(1), PodName: "pod-a"},
		},
		Want: []backendBucket{
			{Key: "pod-a", Source: "pod_name", Count: 1, PercentOfTotal: 50, P50: ms(1), P99: ms(1)},
			{Key: "pod-b", Source: "pod_name", Count: 1, PercentOfTotal: 50, P50: ms(2), P99: ms(2)},
		},
	},
}

func TestAggregate_Backends(t *testing.T) {
	for _, tc := range aggregateBackendsTests {
		t.Run(tc.Name, func(t *testing.T) {
			s := aggregate(tc.Results, time.Second, ConnModelPerWorker)
			require.Len(t, s.Backends, len(tc.Want))
			for i, want := range tc.Want {
				got := s.Backends[i]
				assert.Equal(t, want.Key, got.Key, "Backends[%d].Key", i)
				assert.Equal(t, want.Source, got.Source, "Backends[%d].Source", i)
				assert.Equal(t, want.Count, got.Count, "Backends[%d].Count", i)
				assert.Equal(t, want.ErrorCount, got.ErrorCount, "Backends[%d].ErrorCount", i)
				assert.InDelta(t, want.PercentOfTotal, got.PercentOfTotal, 1e-9, "Backends[%d].PercentOfTotal", i)
				assert.Equal(t, want.P50, got.P50, "Backends[%d].P50", i)
				assert.Equal(t, want.P99, got.P99, "Backends[%d].P99", i)
			}
		})
	}
}

type backendKeyTest struct {
	Name       string
	Result     Result
	WantKey    string
	WantSource string
}

var backendKeyTests = []backendKeyTest{
	{
		Name:       "pod_name wins over hostname and peer",
		Result:     Result{PodName: "echo-abc", PodHostname: "h", PeerAddr: "10.0.0.1:1"},
		WantKey:    "echo-abc",
		WantSource: "pod_name",
	},
	{
		Name:       "hostname used when pod_name empty",
		Result:     Result{PodHostname: "h", PeerAddr: "10.0.0.1:1"},
		WantKey:    "h",
		WantSource: "hostname",
	},
	{
		Name:       "peer used when both kubernetes fields empty",
		Result:     Result{PeerAddr: "10.0.0.1:1"},
		WantKey:    "10.0.0.1:1",
		WantSource: "peer",
	},
	{
		Name:       "all empty falls into unknown bucket",
		Result:     Result{},
		WantKey:    "",
		WantSource: "unknown",
	},
}

func TestBackendKey(t *testing.T) {
	for _, tc := range backendKeyTests {
		t.Run(tc.Name, func(t *testing.T) {
			k, src := backendKey(tc.Result)
			assert.Equal(t, tc.WantKey, k, "key")
			assert.Equal(t, tc.WantSource, src, "source")
		})
	}
}

type computeBackendSkewTest struct {
	Name          string
	Backends      []BackendStats
	WantCountRate float64
	WantP99Rate   float64
}

var computeBackendSkewTests = []computeBackendSkewTest{
	{
		Name:     "empty input has no skew",
		Backends: nil,
	},
	{
		Name:     "single backend has no skew to measure",
		Backends: []BackendStats{{Key: "pod-a", Count: 10, P99: ms(5)}},
	},
	{
		Name: "even counts and equal p99 yield 1.0",
		Backends: []BackendStats{
			{Key: "pod-a", Count: 50, P99: ms(4)},
			{Key: "pod-b", Count: 50, P99: ms(4)},
		},
		WantCountRate: 1.0,
		WantP99Rate:   1.0,
	},
	{
		// 60/40 = 1.5; 12ms / 9ms = 1.333...
		Name: "60/40 split with 9ms vs 12ms p99",
		Backends: []BackendStats{
			{Key: "pod-a", Count: 60, P99: ms(9)},
			{Key: "pod-b", Count: 40, P99: ms(12)},
		},
		WantCountRate: 1.5,
		WantP99Rate:   12.0 / 9.0,
	},
	{
		// pod-c has only errors, so its P99 is zero and it does not
		// contribute to the p99 ratio; it does contribute to count.
		Name: "error-only backend ignored by p99 but counts toward count",
		Backends: []BackendStats{
			{Key: "pod-a", Count: 50, P99: ms(6)},
			{Key: "pod-b", Count: 30, P99: ms(9)},
			{Key: "pod-c", Count: 10, ErrorCount: 10, P99: 0},
		},
		WantCountRate: 5.0,
		WantP99Rate:   9.0 / 6.0,
	},
	{
		Name: "all-error backends report count ratio but no p99 ratio",
		Backends: []BackendStats{
			{Key: "pod-a", Count: 10, ErrorCount: 10, P99: 0},
			{Key: "pod-b", Count: 5, ErrorCount: 5, P99: 0},
		},
		WantCountRate: 2.0,
	},
	{
		// Only one backend has any successful samples; p99 ratio needs
		// at least two qualifying backends.
		Name: "single-success backend leaves p99 ratio undefined",
		Backends: []BackendStats{
			{Key: "pod-a", Count: 50, P99: ms(8)},
			{Key: "pod-b", Count: 50, ErrorCount: 50, P99: 0},
		},
		WantCountRate: 1.0,
	},
}

func TestComputeBackendSkew(t *testing.T) {
	for _, tc := range computeBackendSkewTests {
		t.Run(tc.Name, func(t *testing.T) {
			got := computeBackendSkew(tc.Backends)
			assert.InDelta(t, tc.WantCountRate, got.CountRatio, 1e-9, "CountRatio")
			assert.InDelta(t, tc.WantP99Rate, got.P99Ratio, 1e-9, "P99Ratio")
		})
	}
}

type statusCodesBucket struct {
	Code     codes.Code
	CodeName string
	Count    int
	Messages []ErrorMessageStat
}

type computeStatusCodesTest struct {
	Name    string
	Results []Result
	Want    []statusCodesBucket
}

var computeStatusCodesTests = []computeStatusCodesTest{
	{
		Name:    "empty input produces no buckets",
		Results: nil,
		Want:    nil,
	},
	{
		Name: "successes-only run produces no buckets",
		Results: []Result{
			{TotalDuration: ms(1)},
			{TotalDuration: ms(2)},
		},
		Want: nil,
	},
	{
		Name: "single code aggregates and dedups messages",
		Results: []Result{
			{Err: status.Error(codes.InvalidArgument, "missing query")},
			{Err: status.Error(codes.InvalidArgument, "missing query")},
			{Err: status.Error(codes.InvalidArgument, "bad shape")},
		},
		Want: []statusCodesBucket{
			{
				Code: codes.InvalidArgument, CodeName: "InvalidArgument", Count: 3,
				Messages: []ErrorMessageStat{
					{Message: "missing query", Count: 2},
					{Message: "bad shape", Count: 1},
				},
			},
		},
	},
	{
		Name: "multiple codes sort by count descending",
		Results: []Result{
			{Err: status.Error(codes.Unavailable, "conn reset")},
			{Err: status.Error(codes.Unavailable, "conn reset")},
			{Err: status.Error(codes.Unavailable, "conn reset")},
			{Err: status.Error(codes.InvalidArgument, "bad shape")},
		},
		Want: []statusCodesBucket{
			{
				Code: codes.Unavailable, CodeName: "Unavailable", Count: 3,
				Messages: []ErrorMessageStat{{Message: "conn reset", Count: 3}},
			},
			{
				Code: codes.InvalidArgument, CodeName: "InvalidArgument", Count: 1,
				Messages: []ErrorMessageStat{{Message: "bad shape", Count: 1}},
			},
		},
	},
	{
		// Both buckets carry one error, so the secondary key (Code asc)
		// puts InvalidArgument (3) before Unavailable (14).
		Name: "tied counts sort by code ascending",
		Results: []Result{
			{Err: status.Error(codes.Unavailable, "u")},
			{Err: status.Error(codes.InvalidArgument, "ia")},
		},
		Want: []statusCodesBucket{
			{
				Code: codes.InvalidArgument, CodeName: "InvalidArgument", Count: 1,
				Messages: []ErrorMessageStat{{Message: "ia", Count: 1}},
			},
			{
				Code: codes.Unavailable, CodeName: "Unavailable", Count: 1,
				Messages: []ErrorMessageStat{{Message: "u", Count: 1}},
			},
		},
	},
	{
		// status.Code on a plain error returns Unknown; the wrapped
		// message is preserved so the failure is still legible.
		Name: "non-status errors fall into Unknown",
		Results: []Result{
			{Err: errors.New("acquiring conn: bufconn closed")},
			{Err: errors.New("acquiring conn: bufconn closed")},
		},
		Want: []statusCodesBucket{
			{
				Code: codes.Unknown, CodeName: "Unknown", Count: 2,
				Messages: []ErrorMessageStat{
					{Message: "acquiring conn: bufconn closed", Count: 2},
				},
			},
		},
	},
	{
		// Four distinct messages, only three survive. The dropped one
		// is the lowest-count entry; its count still contributes to the
		// bucket's Count total.
		Name: "top-N keeps three highest-count messages within a code",
		Results: []Result{
			{Err: status.Error(codes.InvalidArgument, "a")},
			{Err: status.Error(codes.InvalidArgument, "a")},
			{Err: status.Error(codes.InvalidArgument, "a")},
			{Err: status.Error(codes.InvalidArgument, "b")},
			{Err: status.Error(codes.InvalidArgument, "b")},
			{Err: status.Error(codes.InvalidArgument, "c")},
			{Err: status.Error(codes.InvalidArgument, "d")},
		},
		Want: []statusCodesBucket{
			{
				Code: codes.InvalidArgument, CodeName: "InvalidArgument", Count: 7,
				Messages: []ErrorMessageStat{
					{Message: "a", Count: 3},
					{Message: "b", Count: 2},
					{Message: "c", Count: 1},
				},
			},
		},
	},
	{
		// Tied counts within a bucket break by Message ascending so the
		// surviving three are deterministic across runs.
		Name: "tied message counts break by message ascending",
		Results: []Result{
			{Err: status.Error(codes.InvalidArgument, "delta")},
			{Err: status.Error(codes.InvalidArgument, "charlie")},
			{Err: status.Error(codes.InvalidArgument, "bravo")},
			{Err: status.Error(codes.InvalidArgument, "alpha")},
		},
		Want: []statusCodesBucket{
			{
				Code: codes.InvalidArgument, CodeName: "InvalidArgument", Count: 4,
				Messages: []ErrorMessageStat{
					{Message: "alpha", Count: 1},
					{Message: "bravo", Count: 1},
					{Message: "charlie", Count: 1},
				},
			},
		},
	},
}

func TestComputeStatusCodes(t *testing.T) {
	for _, tc := range computeStatusCodesTests {
		t.Run(tc.Name, func(t *testing.T) {
			got := computeStatusCodes(tc.Results)
			require.Len(t, got, len(tc.Want))
			for i, want := range tc.Want {
				assert.Equal(t, want.Code, got[i].Code, "Errors[%d].Code", i)
				assert.Equal(t, want.CodeName, got[i].CodeName, "Errors[%d].CodeName", i)
				assert.Equal(t, want.Count, got[i].Count, "Errors[%d].Count", i)
				assert.Equal(t, want.Messages, got[i].TopMessages, "Errors[%d].TopMessages", i)
			}
		})
	}
}

func TestComputeStatusCodes_TruncatesLongMessages(t *testing.T) {
	long := strings.Repeat("x", maxErrorMessageLen+10)
	got := computeStatusCodes([]Result{
		{Err: status.Error(codes.InvalidArgument, long)},
	})
	require.Len(t, got, 1)
	require.Len(t, got[0].TopMessages, 1)
	want := strings.Repeat("x", maxErrorMessageLen) + "..."
	assert.Equal(t, want, got[0].TopMessages[0].Message)
}

func TestTruncateMessage_RuneBoundary(t *testing.T) {
	// A two-byte rune ("é") repeated keeps the cut on a rune boundary;
	// the result must remain valid UTF-8 even when the source contains
	// multi-byte runes.
	long := strings.Repeat("é", maxErrorMessageLen+5)
	got := truncateMessage(long)
	assert.Equal(t, maxErrorMessageLen+len("..."), len([]rune(got)))
	assert.True(t, strings.HasSuffix(got, "..."), "truncated string ends with ellipsis")
}

func ascendingResults(n int) []Result {
	rs := make([]Result, n)
	for i := range rs {
		rs[i] = Result{
			TotalDuration:    ms(i + 1),
			ServerDurationNs: 0,
		}
	}
	return rs
}

func ascendingDurations(n int) []time.Duration {
	ds := make([]time.Duration, n)
	for i := range ds {
		ds[i] = ms(i + 1)
	}
	return ds
}
