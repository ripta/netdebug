package bench

import (
	"bytes"
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

func newBufconnEchoServer(t *testing.T) *bufconn.Listener {
	t.Helper()

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	echov1.RegisterEchoerServer(gs, &echov1.Server{})

	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("grpc serve returned: %v", err)
		}
	}()
	t.Cleanup(gs.Stop)

	return lis
}

func bufconnDialOpts(lis *bufconn.Listener) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
	}
}

func TestRun_BufconnSmoke(t *testing.T) {
	lis := newBufconnEchoServer(t)

	var buf bytes.Buffer
	cfg := &Config{
		Target:       "passthrough://bufnet",
		Plaintext:    true,
		Concurrency:  2,
		Duration:     200 * time.Millisecond,
		Payload:      defaultMix,
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatHuman,
		Output:       &buf,
		dialOpts:     bufconnDialOpts(lis),
	}

	start := time.Now()
	require.NoError(t, cfg.Run(context.Background()))
	wall := time.Since(start)

	assert.Less(t, wall, 300*time.Millisecond, "run should terminate within 1.5x duration")

	out := buf.String()
	for _, want := range []string{
		"Target:      passthrough://bufnet",
		"Concurrency: 2",
		"Duration:    200ms",
		"Conn model:  per-worker",
		"Requests:",
		"Errors:      0",
		"Throughput:",
		"Latency (total):",
		"Latency (server):",
		"Latency (network):",
		"p50:",
		"p99:",
	} {
		assert.Contains(t, out, want)
	}
}

// newBufconnErrorServer stands up an echo server whose Echo handler is
// short-circuited by a unary server interceptor returning the given
// status. The interceptor sits in front of the real Echoer, so the
// handler never runs; the bench client sees nothing but the forced
// status, which is what TestRun_ForcedErrors_GroupedByCode needs.
func newBufconnErrorServer(t *testing.T, code codes.Code, msg string) *bufconn.Listener {
	t.Helper()

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer(grpc.UnaryInterceptor(
		func(_ context.Context, _ any, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
			return nil, status.Error(code, msg)
		},
	))
	echov1.RegisterEchoerServer(gs, &echov1.Server{})

	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("grpc serve returned: %v", err)
		}
	}()
	t.Cleanup(gs.Stop)

	return lis
}

func TestRun_ForcedErrors_GroupedByCode(t *testing.T) {
	lis := newBufconnErrorServer(t, codes.InvalidArgument, "boom")

	cfg := &Config{
		Target:       "passthrough://bufnet",
		Plaintext:    true,
		Concurrency:  2,
		Duration:     200 * time.Millisecond,
		Payload:      defaultMix,
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatHuman,
		Output:       io.Discard,
		dialOpts:     bufconnDialOpts(lis),
	}

	s, err := cfg.run(context.Background())
	require.NoError(t, err)

	assert.Greater(t, s.Count, 0, "should issue at least one request")
	assert.Equal(t, s.Count, s.ErrorCount, "every request was forced to fail")

	require.Len(t, s.Errors, 1, "single forced code should produce one bucket")
	got := s.Errors[0]
	assert.Equal(t, codes.InvalidArgument, got.Code)
	assert.Equal(t, "InvalidArgument", got.CodeName)
	assert.Equal(t, s.ErrorCount, got.Count, "bucket counts every errored result")
	require.Len(t, got.TopMessages, 1)
	assert.Equal(t, "boom", got.TopMessages[0].Message)
	assert.Equal(t, s.ErrorCount, got.TopMessages[0].Count)
}

// newBufconnUpstreamHeaderServer stands up an echo server whose unary
// interceptor injects a synthetic x-envoy-upstream-service-time
// response header before delegating to the real handler. Modelled on
// newBufconnErrorServer so the test mirrors the established harness.
func newBufconnUpstreamHeaderServer(t *testing.T, ms int) *bufconn.Listener {
	t.Helper()

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer(grpc.UnaryInterceptor(
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
			if err := grpc.SetHeader(ctx, metadata.Pairs(headerEnvoyUpstreamTime, strconv.Itoa(ms))); err != nil {
				t.Logf("setting envoy header: %v", err)
			}
			return h(ctx, req)
		},
	))
	echov1.RegisterEchoerServer(gs, &echov1.Server{})

	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("grpc serve returned: %v", err)
		}
	}()
	t.Cleanup(gs.Stop)

	return lis
}

func TestRun_CapturesEnvoyUpstreamHeader(t *testing.T) {
	const upstreamMs = 7
	lis := newBufconnUpstreamHeaderServer(t, upstreamMs)

	cfg := &Config{
		Target:       "passthrough://bufnet",
		Plaintext:    true,
		Concurrency:  2,
		Duration:     200 * time.Millisecond,
		Payload:      defaultMix,
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatHuman,
		Output:       io.Discard,
		dialOpts:     bufconnDialOpts(lis),
	}

	s, err := cfg.run(context.Background())
	require.NoError(t, err)

	assert.Greater(t, s.Count, 0, "should issue at least one request")
	assert.Equal(t, 0, s.ErrorCount)
	want := time.Duration(upstreamMs) * time.Millisecond
	assert.Equal(t, s.Count, s.Upstream.Count, "every successful request should carry the header")
	assert.Equal(t, want, s.Upstream.Min)
	assert.Equal(t, want, s.Upstream.Max)
	assert.Equal(t, want, s.Upstream.P50)
}

// newBufconnHeaderCaptureServer stands up an echo server whose unary
// interceptor records every call's incoming metadata values for key into
// a thread-safe slice. The bench client's --header injection should land
// in metadata.FromIncomingContext on the server side.
func newBufconnHeaderCaptureServer(t *testing.T, key string) (*bufconn.Listener, *headerSink) {
	t.Helper()

	sink := &headerSink{}
	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer(grpc.UnaryInterceptor(
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, h grpc.UnaryHandler) (any, error) {
			md, _ := metadata.FromIncomingContext(ctx)
			sink.add(md.Get(key))
			return h(ctx, req)
		},
	))
	echov1.RegisterEchoerServer(gs, &echov1.Server{})

	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("grpc serve returned: %v", err)
		}
	}()
	t.Cleanup(gs.Stop)

	return lis, sink
}

type headerSink struct {
	mu     sync.Mutex
	values [][]string
}

func (h *headerSink) add(vals []string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	dup := append([]string(nil), vals...)
	h.values = append(h.values, dup)
}

func (h *headerSink) snapshot() [][]string {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([][]string, len(h.values))
	for i, v := range h.values {
		out[i] = append([]string(nil), v...)
	}
	return out
}

func TestRun_AttachesCustomHeaders(t *testing.T) {
	const key = "x-test-route"
	const want = "alpha"
	lis, sink := newBufconnHeaderCaptureServer(t, key)

	cfg := &Config{
		Target:       "passthrough://bufnet",
		Plaintext:    true,
		Concurrency:  2,
		Duration:     200 * time.Millisecond,
		Payload:      defaultMix,
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatHuman,
		Headers:      map[string]string{key: want},
		Output:       io.Discard,
		dialOpts:     bufconnDialOpts(lis),
	}

	s, err := cfg.run(context.Background())
	require.NoError(t, err)
	assert.Greater(t, s.Count, 0, "should issue at least one request")
	assert.Equal(t, 0, s.ErrorCount)

	calls := sink.snapshot()
	require.GreaterOrEqual(t, len(calls), s.Count,
		"interceptor should see every call the client recorded (plus any late cancellations)")
	for i, vals := range calls {
		require.Len(t, vals, 1, "call %d: header %q should be set exactly once", i, key)
		assert.Equal(t, want, vals[0], "call %d: header %q value", i, key)
	}
}

func TestRun_BufconnPopulatesSummary(t *testing.T) {
	lis := newBufconnEchoServer(t)

	cfg := &Config{
		Target:       "passthrough://bufnet",
		Plaintext:    true,
		Concurrency:  2,
		Duration:     200 * time.Millisecond,
		Payload:      defaultMix,
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatHuman,
		Output:       io.Discard,
		dialOpts:     bufconnDialOpts(lis),
	}

	s, err := cfg.run(context.Background())
	require.NoError(t, err)

	assert.Greater(t, s.Count, 0, "should issue at least one request")
	assert.Equal(t, 0, s.ErrorCount)
	assert.Equal(t, s.Count, s.Total.Count)
	assert.Greater(t, s.Total.P50, time.Duration(0))
	assert.Greater(t, s.Total.P99, time.Duration(0))
	assert.Equal(t, s.Count, s.Server.Count)
	assert.Greater(t, s.Server.P50, time.Duration(0))
	assert.Equal(t, s.Count, s.Network.Count)
	assert.Greater(t, s.Throughput, 0.0)
	assert.Equal(t, ConnModelPerWorker, s.ConnModel)

	// The bufconn echo server runs without an HTTP wrapper, so no
	// KubernetesInfo is attached to the context and rsp.Kubernetes comes
	// back zero-valued. Per-backend grouping must still fire via the peer
	// fallback for every recorded request.
	require.NotEmpty(t, s.Backends)
	var totalBackendCount int
	for _, b := range s.Backends {
		assert.Equal(t, "peer", b.Source, "backend %q should fall through to peer", b.Key)
		assert.NotEmpty(t, b.Key, "peer-sourced backend must carry an address")
		totalBackendCount += b.Count
	}
	assert.Equal(t, s.Count, totalBackendCount, "every recorded request appears in exactly one bucket")
}
