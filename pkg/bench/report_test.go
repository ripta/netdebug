package bench

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteReport_PopulatedSummary(t *testing.T) {
	cfg := &Config{
		Target:      "127.0.0.1:9999",
		Concurrency: 4,
		Duration:    5 * time.Second,
	}
	s := Summary{
		Count:      100,
		ErrorCount: 2,
		Elapsed:    5 * time.Second,
		Throughput: 20,
		ConnModel:  ConnModelPerWorker,
		Total: LatencyStats{
			Count: 98, Min: time.Millisecond, Mean: 4 * time.Millisecond,
			Stddev: 2 * time.Millisecond, Max: 20 * time.Millisecond,
			P50: 3 * time.Millisecond, P90: 8 * time.Millisecond, P99: 12 * time.Millisecond,
		},
		Server: LatencyStats{
			Count: 98, Min: 500 * time.Microsecond, Mean: 2 * time.Millisecond,
			Stddev: time.Millisecond, Max: 10 * time.Millisecond,
			P50: 2 * time.Millisecond, P90: 5 * time.Millisecond, P99: 8 * time.Millisecond,
		},
		Network: LatencyStats{
			Count: 98, Min: 500 * time.Microsecond, Mean: 2 * time.Millisecond,
			Stddev: time.Millisecond, Max: 10 * time.Millisecond,
			P50: time.Millisecond, P90: 3 * time.Millisecond, P99: 4 * time.Millisecond,
		},
	}

	var buf bytes.Buffer
	require.NoError(t, writeReport(&buf, cfg, s))

	out := buf.String()
	for _, want := range []string{
		"Target:      127.0.0.1:9999",
		"Concurrency: 4",
		"Duration:    5s",
		"Conn model:  per-worker",
		"Requests:    100",
		"Errors:      2",
		"Elapsed:     5s",
		"Throughput:  20.00 req/s",
		"Latency (total):",
		"Latency (server):",
		"Latency (network):",
		"count:  98",
		"min:    1ms",
		"mean:   4ms",
		"stddev: 2ms",
		"p50:    3ms",
		"p90:    8ms",
		"p99:    12ms",
		"max:    20ms",
	} {
		assert.Contains(t, out, want)
	}
}

func TestWriteReport_NoSuccessesRendersNA(t *testing.T) {
	cfg := &Config{Target: "127.0.0.1:9999", Concurrency: 1, Duration: time.Second}
	s := Summary{Count: 3, ErrorCount: 3, Elapsed: time.Second, Throughput: 3}

	var buf bytes.Buffer
	require.NoError(t, writeReport(&buf, cfg, s))

	out := buf.String()
	assert.Contains(t, out, "Latency (total): n/a")
	assert.Contains(t, out, "Latency (server): n/a")
	assert.Contains(t, out, "Latency (network): n/a")
	assert.Contains(t, out, "Errors:      3")
}

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("writer is broken")
}

func TestWriteReport_PropagatesWriterError(t *testing.T) {
	cfg := &Config{Target: "x", Concurrency: 1, Duration: time.Second}
	s := Summary{Count: 1, Elapsed: time.Second, Throughput: 1}
	assert.Error(t, writeReport(failingWriter{}, cfg, s))
}
