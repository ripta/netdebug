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
		LatencyP50: 3 * time.Millisecond,
		LatencyP99: 12 * time.Millisecond,
		ConnModel:  ConnModelPerWorker,
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
		"Latency p50: 3ms",
		"Latency p99: 12ms",
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
	assert.Contains(t, out, "Latency p50: n/a")
	assert.Contains(t, out, "Latency p99: n/a")
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
