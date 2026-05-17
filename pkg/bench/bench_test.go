package bench

import (
	"bytes"
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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
		Target:      "passthrough://bufnet",
		Plaintext:   true,
		Concurrency: 2,
		Duration:    200 * time.Millisecond,
		Payload:     defaultMix,
		Compression: CompressionIdentity,
		ConnModel:   ConnModelPerWorker,
		Output:      &buf,
		dialOpts:    bufconnDialOpts(lis),
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

func TestRun_BufconnPopulatesSummary(t *testing.T) {
	lis := newBufconnEchoServer(t)

	cfg := &Config{
		Target:      "passthrough://bufnet",
		Plaintext:   true,
		Concurrency: 2,
		Duration:    200 * time.Millisecond,
		Payload:     defaultMix,
		Compression: CompressionIdentity,
		ConnModel:   ConnModelPerWorker,
		Output:      io.Discard,
		dialOpts:    bufconnDialOpts(lis),
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
}
