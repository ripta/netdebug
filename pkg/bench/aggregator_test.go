package bench

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
