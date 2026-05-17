package bench

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
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
		Upstream: LatencyStats{
			Count: 50, Min: time.Millisecond, Mean: 3 * time.Millisecond,
			Stddev: time.Millisecond, Max: 9 * time.Millisecond,
			P50: 3 * time.Millisecond, P90: 6 * time.Millisecond, P99: 8 * time.Millisecond,
		},
		Backends: []BackendStats{
			{Key: "pod-a", Source: "pod_name", Count: 60, PercentOfTotal: 60, P50: 3 * time.Millisecond, P99: 9 * time.Millisecond, ErrorCount: 1},
			{Key: "pod-b", Source: "pod_name", Count: 40, PercentOfTotal: 40, P50: 4 * time.Millisecond, P99: 12 * time.Millisecond, ErrorCount: 1},
		},
		BackendSkew: BackendSkew{CountRatio: 1.5, P99Ratio: 12.0 / 9.0},
		Errors: []StatusCodeStats{
			{
				Code: codes.InvalidArgument, CodeName: "InvalidArgument", Count: 2,
				TopMessages: []ErrorMessageStat{
					{Message: "missing query", Count: 1},
					{Message: "bad shape", Count: 1},
				},
			},
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
		"Latency (upstream):",
		"count:  50",
		"count:  98",
		"min:    1ms",
		"mean:   4ms",
		"stddev: 2ms",
		"p50:    3ms",
		"p90:    8ms",
		"p99:    12ms",
		"max:    20ms",
		"Backends:",
		"pod_name=pod-a (60 req, 60.0%): p50=3ms p99=9ms errors=1",
		"pod_name=pod-b (40 req, 40.0%): p50=4ms p99=12ms errors=1",
		"Backend skew: count=1.50 p99=1.33",
		"Errors by code:",
		"InvalidArgument (2 req): \"missing query\" (1), \"bad shape\" (1)",
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
	assert.Contains(t, out, "Latency (upstream): n/a")
	assert.Contains(t, out, "Backends: n/a")
	assert.Contains(t, out, "Backend skew: count=n/a p99=n/a")
	assert.Contains(t, out, "Errors:      3")
	assert.Contains(t, out, "Errors by code: n/a")
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

func TestWriteJSONReport_Smoke(t *testing.T) {
	cfg := &Config{
		Target:       "127.0.0.1:9999",
		Plaintext:    true,
		Concurrency:  4,
		Duration:     5 * time.Second,
		Payload:      defaultMix,
		EmbeddingDim: 1024,
		BytesSize:    1024,
		StringLen:    1024,
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatJSON,
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
		Backends: []BackendStats{
			{Key: "pod-a", Source: "pod_name", Count: 60, PercentOfTotal: 60, P50: 3 * time.Millisecond, P99: 9 * time.Millisecond, ErrorCount: 1},
		},
		BackendSkew: BackendSkew{CountRatio: 1.5, P99Ratio: 1.33},
		Errors: []StatusCodeStats{
			{
				Code: codes.InvalidArgument, CodeName: "InvalidArgument", Count: 2,
				TopMessages: []ErrorMessageStat{{Message: "boom", Count: 2}},
			},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, writeJSONReport(&buf, cfg, s))

	// Output must be a single valid JSON object.
	var root map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &root))

	cfgOut, ok := root["config"].(map[string]any)
	require.True(t, ok, "config block must be present and an object")
	assert.Equal(t, "127.0.0.1:9999", cfgOut["target"])
	assert.Equal(t, float64(4), cfgOut["concurrency"])
	assert.Equal(t, "5s", cfgOut["duration"], "durations render in Go string form")
	assert.Equal(t, "identity", cfgOut["compression"])
	assert.Equal(t, "per-worker", cfgOut["conn_model"])

	payload, ok := cfgOut["payload"].([]any)
	require.True(t, ok, "payload must be an array")
	require.Len(t, payload, 1)
	entry, ok := payload[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "embedding-float", entry["shape"], "shapes render as flag-facing names")
	assert.Equal(t, float64(1), entry["weight"])

	sumOut, ok := root["summary"].(map[string]any)
	require.True(t, ok, "summary block must be present and an object")
	assert.Equal(t, float64(100), sumOut["count"])
	assert.Equal(t, float64(2), sumOut["error_count"])
	assert.Equal(t, "5s", sumOut["elapsed"])
	assert.Equal(t, float64(20), sumOut["throughput_rps"])

	total, ok := sumOut["total"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "3ms", total["p50"])
	assert.Equal(t, "12ms", total["p99"])

	backends, ok := sumOut["backends"].([]any)
	require.True(t, ok)
	require.Len(t, backends, 1)
	b, ok := backends[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "pod-a", b["key"])
	assert.Equal(t, "pod_name", b["source"])

	errs, ok := sumOut["errors"].([]any)
	require.True(t, ok)
	require.Len(t, errs, 1)
	e, ok := errs[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "InvalidArgument", e["code_name"])
	assert.Equal(t, float64(codes.InvalidArgument), e["code"])
}

func TestWriteJSONReport_UnknownPayloadShape(t *testing.T) {
	cfg := &Config{
		Target:       "127.0.0.1:9999",
		Concurrency:  1,
		Duration:     time.Second,
		Payload:      PayloadMix{{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_UNSPECIFIED, Weight: 1}},
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatJSON,
	}
	var buf bytes.Buffer
	require.NoError(t, writeJSONReport(&buf, cfg, Summary{}))

	var root map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &root))
	cfgOut := root["config"].(map[string]any)
	payload := cfgOut["payload"].([]any)
	entry := payload[0].(map[string]any)
	// Unknown shapes fall back to the proto-generated enum name so output
	// stays self-describing instead of dropping to a raw integer.
	assert.Equal(t, "PAYLOAD_SHAPE_UNSPECIFIED", entry["shape"])
}

func TestWriteJSONReport_PropagatesWriterError(t *testing.T) {
	cfg := &Config{
		Target: "x", Concurrency: 1, Duration: time.Second,
		Payload: defaultMix, Compression: CompressionIdentity,
		ConnModel: ConnModelPerWorker, OutputFormat: OutputFormatJSON,
	}
	assert.Error(t, writeJSONReport(failingWriter{}, cfg, Summary{}))
}
