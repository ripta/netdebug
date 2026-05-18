package bench

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
		Labels:      map[string]string{"mesh": "istio", "payload": "embedding-float"},
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
		"Labels:      mesh=istio payload=embedding-float",
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
	assert.NotContains(t, out, "Labels:", "header should omit the labels line when no labels are set")
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
		Labels:       map[string]string{"mesh": "istio", "payload": "embedding-float"},
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

	labels, ok := cfgOut["labels"].(map[string]any)
	require.True(t, ok, "labels must be a JSON object")
	assert.Equal(t, "istio", labels["mesh"])
	assert.Equal(t, "embedding-float", labels["payload"])

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

	// A nil Labels map must still serialize as an empty object, not null,
	// so consumers can index into config.labels unconditionally.
	labels, ok := cfgOut["labels"].(map[string]any)
	require.True(t, ok, "labels must be a JSON object even when no labels were set")
	assert.Empty(t, labels)
}

func TestWriteJSONReport_PropagatesWriterError(t *testing.T) {
	cfg := &Config{
		Target: "x", Concurrency: 1, Duration: time.Second,
		Payload: defaultMix, Compression: CompressionIdentity,
		ConnModel: ConnModelPerWorker, OutputFormat: OutputFormatJSON,
	}
	assert.Error(t, writeJSONReport(failingWriter{}, cfg, Summary{}))
}

type jsonDurationUnmarshalTest struct {
	Name    string
	Input   string
	Want    time.Duration
	WantErr bool
}

var jsonDurationUnmarshalTests = []jsonDurationUnmarshalTest{
	{Name: "seconds", Input: `"5s"`, Want: 5 * time.Second},
	{Name: "milliseconds", Input: `"300ms"`, Want: 300 * time.Millisecond},
	{Name: "compound", Input: `"1h30m"`, Want: 90 * time.Minute},
	{Name: "zero string", Input: `"0s"`, Want: 0},
	{Name: "empty string is zero", Input: `""`, Want: 0},
	{Name: "malformed", Input: `"not a duration"`, WantErr: true},
	{Name: "wrong json type", Input: `123`, WantErr: true},
}

func TestJSONDuration_Unmarshal(t *testing.T) {
	for _, tc := range jsonDurationUnmarshalTests {
		t.Run(tc.Name, func(t *testing.T) {
			var d jsonDuration
			err := json.Unmarshal([]byte(tc.Input), &d)
			if tc.WantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.Want, time.Duration(d))
		})
	}
}

type jsonPayloadShapeUnmarshalTest struct {
	Name    string
	Input   string
	Want    echov1.PayloadShape
	WantErr bool
}

var jsonPayloadShapeUnmarshalTests = []jsonPayloadShapeUnmarshalTest{
	{Name: "kebab string", Input: `"string"`, Want: echov1.PayloadShape_PAYLOAD_SHAPE_STRING},
	{Name: "kebab bytes", Input: `"bytes"`, Want: echov1.PayloadShape_PAYLOAD_SHAPE_BYTES},
	{Name: "kebab embedding-float", Input: `"embedding-float"`, Want: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT},
	{Name: "kebab embedding-bytes", Input: `"embedding-bytes"`, Want: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES},
	{Name: "kebab mixed", Input: `"mixed"`, Want: echov1.PayloadShape_PAYLOAD_SHAPE_MIXED},
	{Name: "proto unspecified fallback", Input: `"PAYLOAD_SHAPE_UNSPECIFIED"`, Want: echov1.PayloadShape_PAYLOAD_SHAPE_UNSPECIFIED},
	{Name: "proto named fallback", Input: `"PAYLOAD_SHAPE_EMBEDDING_FLOAT"`, Want: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT},
	{Name: "unknown rejected", Input: `"banana"`, WantErr: true},
	{Name: "wrong json type", Input: `42`, WantErr: true},
}

func TestJSONPayloadShape_Unmarshal(t *testing.T) {
	for _, tc := range jsonPayloadShapeUnmarshalTests {
		t.Run(tc.Name, func(t *testing.T) {
			var p jsonPayloadShape
			err := json.Unmarshal([]byte(tc.Input), &p)
			if tc.WantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.Want, echov1.PayloadShape(p))
		})
	}
}

// fullReportFixture returns a Config and Summary with every documented
// field populated. Round-trip and schema-stability tests share the
// fixture so both exercise the same surface.
func fullReportFixture() (*Config, Summary) {
	cfg := &Config{
		Target:      "127.0.0.1:9999",
		Plaintext:   true,
		Concurrency: 4,
		Duration:    5 * time.Second,
		Payload: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 3},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 1},
		},
		EmbeddingDim: 1024,
		BytesSize:    2048,
		StringLen:    512,
		Compression:  CompressionGzip,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatJSON,
		Labels: map[string]string{
			"mesh":    "istio",
			"payload": "embedding-float",
			"run":     "a",
		},
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
			{Key: "pod-a", Source: "pod_name", Count: 60, ErrorCount: 1, PercentOfTotal: 60, P50: 3 * time.Millisecond, P99: 9 * time.Millisecond},
			{Key: "pod-b", Source: "hostname", Count: 40, ErrorCount: 1, PercentOfTotal: 40, P50: 4 * time.Millisecond, P99: 12 * time.Millisecond},
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
			{
				Code: codes.Unavailable, CodeName: "Unavailable", Count: 1,
				TopMessages: []ErrorMessageStat{
					{Message: "upstream dropped", Count: 1},
				},
			},
		},
	}
	return cfg, s
}

func TestWriteJSONReport_RoundTrip(t *testing.T) {
	cfg, s := fullReportFixture()

	var buf bytes.Buffer
	require.NoError(t, writeJSONReport(&buf, cfg, s))

	var got jsonReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))

	// The wire form is what newJSONReport produces; the round-trip must
	// reproduce that exact value without losing or transforming any field.
	want := newJSONReport(cfg, s)
	require.Equal(t, want, got)
}

// assertObjectKeys fails the test if `got` is missing any key in
// `expected` or carries any key not in `expected`. Both kinds of mismatch
// are reported separately so a schema diff is obvious.
func assertObjectKeys(t *testing.T, path string, got map[string]any, expected []string) {
	t.Helper()
	want := make(map[string]struct{}, len(expected))
	for _, k := range expected {
		want[k] = struct{}{}
	}
	for k := range want {
		_, ok := got[k]
		assert.True(t, ok, "missing key %q at %s", k, path)
	}
	for k := range got {
		_, ok := want[k]
		assert.True(t, ok, "unexpected key %q at %s", k, path)
	}
}

func TestWriteJSONReport_SchemaStability(t *testing.T) {
	cfg, s := fullReportFixture()

	var buf bytes.Buffer
	require.NoError(t, writeJSONReport(&buf, cfg, s))

	var root map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &root))

	assertObjectKeys(t, "$", root, []string{"config", "summary"})

	cfgKeys := []string{
		"target", "plaintext", "concurrency", "duration", "payload",
		"embedding_dim", "bytes_size", "string_len", "compression",
		"conn_model", "labels",
	}
	cfgObj := root["config"].(map[string]any)
	assertObjectKeys(t, "$.config", cfgObj, cfgKeys)

	payload := cfgObj["payload"].([]any)
	require.NotEmpty(t, payload)
	for i, entry := range payload {
		assertObjectKeys(t, fmt.Sprintf("$.config.payload[%d]", i),
			entry.(map[string]any), []string{"shape", "weight"})
	}

	sumKeys := []string{
		"count", "error_count", "elapsed", "throughput_rps", "conn_model",
		"total", "server", "network", "upstream",
		"backends", "backend_skew", "errors",
	}
	sumObj := root["summary"].(map[string]any)
	assertObjectKeys(t, "$.summary", sumObj, sumKeys)

	latencyKeys := []string{"count", "min", "mean", "stddev", "max", "p50", "p90", "p99"}
	for _, name := range []string{"total", "server", "network", "upstream"} {
		block, ok := sumObj[name].(map[string]any)
		require.True(t, ok, "summary.%s must be an object", name)
		assertObjectKeys(t, "$.summary."+name, block, latencyKeys)
	}

	backendKeys := []string{"key", "source", "count", "error_count", "percent_of_total", "p50", "p99"}
	backends := sumObj["backends"].([]any)
	require.NotEmpty(t, backends)
	for i, b := range backends {
		assertObjectKeys(t, fmt.Sprintf("$.summary.backends[%d]", i),
			b.(map[string]any), backendKeys)
	}

	skew := sumObj["backend_skew"].(map[string]any)
	assertObjectKeys(t, "$.summary.backend_skew", skew, []string{"count_ratio", "p99_ratio"})

	errs := sumObj["errors"].([]any)
	require.NotEmpty(t, errs)
	for i, e := range errs {
		entry := e.(map[string]any)
		assertObjectKeys(t, fmt.Sprintf("$.summary.errors[%d]", i),
			entry, []string{"code", "code_name", "count", "top_messages"})
		msgs := entry["top_messages"].([]any)
		require.NotEmpty(t, msgs)
		for j, m := range msgs {
			assertObjectKeys(t, fmt.Sprintf("$.summary.errors[%d].top_messages[%d]", i, j),
				m.(map[string]any), []string{"message", "count"})
		}
	}
}

type labelsPassThroughTest struct {
	Name   string
	Labels map[string]string
	Want   map[string]string
}

var labelsPassThroughTests = []labelsPassThroughTest{
	{
		Name:   "multi-key",
		Labels: map[string]string{"mesh": "istio", "payload": "embedding-float", "run": "a"},
		Want:   map[string]string{"mesh": "istio", "payload": "embedding-float", "run": "a"},
	},
	{
		Name:   "special characters in values",
		Labels: map[string]string{"note": "before=after spaces", "unicode": "naïve résumé", "empty": ""},
		Want:   map[string]string{"note": "before=after spaces", "unicode": "naïve résumé", "empty": ""},
	},
	{
		Name:   "nil renders empty object",
		Labels: nil,
		Want:   map[string]string{},
	},
	{
		Name:   "empty map renders empty object",
		Labels: map[string]string{},
		Want:   map[string]string{},
	},
}

func TestWriteJSONReport_LabelsPassThrough(t *testing.T) {
	for _, tc := range labelsPassThroughTests {
		t.Run(tc.Name, func(t *testing.T) {
			cfg := &Config{
				Target:       "127.0.0.1:9999",
				Concurrency:  1,
				Duration:     time.Second,
				Payload:      defaultMix,
				Compression:  CompressionIdentity,
				ConnModel:    ConnModelPerWorker,
				OutputFormat: OutputFormatJSON,
				Labels:       tc.Labels,
			}

			var buf bytes.Buffer
			require.NoError(t, writeJSONReport(&buf, cfg, Summary{}))

			var got jsonReport
			require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
			assert.Equal(t, tc.Want, got.Config.Labels)
		})
	}
}
