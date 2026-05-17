package bench

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type summarizeTest struct {
	Name       string
	Results    []Result
	Elapsed    time.Duration
	WantCount  int
	WantErrors int
	WantP50    time.Duration
	WantP99    time.Duration
	WantThrPS  float64
}

var summarizeTests = []summarizeTest{
	{
		Name:       "empty input",
		Results:    nil,
		Elapsed:    time.Second,
		WantCount:  0,
		WantErrors: 0,
		WantP50:    0,
		WantP99:    0,
		WantThrPS:  0,
	},
	{
		Name: "single success",
		Results: []Result{
			{TotalDuration: 5 * time.Millisecond},
		},
		Elapsed:    time.Second,
		WantCount:  1,
		WantErrors: 0,
		WantP50:    5 * time.Millisecond,
		WantP99:    5 * time.Millisecond,
		WantThrPS:  1,
	},
	{
		Name: "errors-only run leaves latencies zero",
		Results: []Result{
			{Err: errors.New("boom")},
			{Err: errors.New("boom")},
			{Err: errors.New("boom")},
		},
		Elapsed:    time.Second,
		WantCount:  3,
		WantErrors: 3,
		WantP50:    0,
		WantP99:    0,
		WantThrPS:  3,
	},
	{
		Name: "mixed errors and successes ignore errors in percentiles",
		Results: []Result{
			{TotalDuration: 1 * time.Millisecond},
			{Err: errors.New("boom")},
			{TotalDuration: 3 * time.Millisecond},
			{Err: errors.New("boom")},
			{TotalDuration: 5 * time.Millisecond},
		},
		Elapsed:    time.Second,
		WantCount:  5,
		WantErrors: 2,
		WantP50:    3 * time.Millisecond,
		WantP99:    5 * time.Millisecond,
		WantThrPS:  5,
	},
	{
		Name:       "100 ascending samples pick nearest rank",
		Results:    ascendingResults(100),
		Elapsed:    time.Second,
		WantCount:  100,
		WantErrors: 0,
		WantP50:    50 * time.Millisecond,
		WantP99:    99 * time.Millisecond,
		WantThrPS:  100,
	},
	{
		Name:       "zero elapsed produces zero throughput",
		Results:    []Result{{TotalDuration: time.Millisecond}},
		Elapsed:    0,
		WantCount:  1,
		WantErrors: 0,
		WantP50:    time.Millisecond,
		WantP99:    time.Millisecond,
		WantThrPS:  0,
	},
}

func TestSummarize(t *testing.T) {
	for _, tc := range summarizeTests {
		t.Run(tc.Name, func(t *testing.T) {
			s := summarize(tc.Results, tc.Elapsed, ConnModelPerWorker)
			assert.Equal(t, tc.WantCount, s.Count, "Count")
			assert.Equal(t, tc.WantErrors, s.ErrorCount, "ErrorCount")
			assert.Equal(t, tc.WantP50, s.LatencyP50, "LatencyP50")
			assert.Equal(t, tc.WantP99, s.LatencyP99, "LatencyP99")
			assert.InDelta(t, tc.WantThrPS, s.Throughput, 1e-9, "Throughput")
			assert.Equal(t, tc.Elapsed, s.Elapsed, "Elapsed")
			assert.Equal(t, ConnModelPerWorker, s.ConnModel, "ConnModel")
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
	{Name: "p50 of 100 is rank 50", Sorted: ascendingDurations(100), P: 50, Want: 50 * time.Millisecond},
	{Name: "p99 of 100 is rank 99", Sorted: ascendingDurations(100), P: 99, Want: 99 * time.Millisecond},
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
		rs[i] = Result{TotalDuration: time.Duration(i+1) * time.Millisecond}
	}
	return rs
}

func ascendingDurations(n int) []time.Duration {
	ds := make([]time.Duration, n)
	for i := range ds {
		ds[i] = time.Duration(i+1) * time.Millisecond
	}
	return ds
}
