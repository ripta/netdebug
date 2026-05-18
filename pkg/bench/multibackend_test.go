package bench

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

// echoBackends is the multi-backend test harness used by the conn-model
// per-mode behavior tests. Each backend has its own bufconn listener and
// its own request counter, incremented by a unary server interceptor.
// dialOpts returns a context dialer that rotates across the backends on
// each call, which combined with grpc-go's default pick_first LB pins one
// ClientConn to whichever backend the dialer handed it. The rotation lives
// at dial time, not in an LB policy, so a single shared ClientConn pins to
// one backend rather than balancing per call.
type echoBackends struct {
	listeners []*bufconn.Listener
	requests  []atomic.Int64
	dials     atomic.Int64
}

// newBufconnEchoBackends stands up two in-process echo backends; the bench
// conn-model tests have no need for more than two, so the count is fixed.
func newBufconnEchoBackends(t *testing.T) *echoBackends {
	t.Helper()

	const n = 2
	be := &echoBackends{
		listeners: make([]*bufconn.Listener, n),
		requests:  make([]atomic.Int64, n),
	}
	for i := 0; i < n; i++ {
		lis := bufconn.Listen(1 << 20)
		gs := grpc.NewServer(grpc.UnaryInterceptor(countingInterceptor(&be.requests[i])))
		echov1.RegisterEchoerServer(gs, &echov1.Server{})

		go func() {
			if err := gs.Serve(lis); err != nil {
				t.Logf("grpc serve returned: %v", err)
			}
		}()
		t.Cleanup(gs.Stop)

		be.listeners[i] = lis
	}
	return be
}

func countingInterceptor(c *atomic.Int64) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		c.Add(1)
		return handler(ctx, req)
	}
}

func (be *echoBackends) dialOpts() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			idx := int(be.dials.Add(1)-1) % len(be.listeners)
			return be.listeners[idx].DialContext(ctx)
		}),
	}
}

func (be *echoBackends) dialer() dialFunc {
	return func() (*grpc.ClientConn, error) {
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		opts = append(opts, be.dialOpts()...)
		return grpc.NewClient("passthrough://bufnet", opts...)
	}
}

func (be *echoBackends) dialCount() int64 { return be.dials.Load() }

func (be *echoBackends) requestCount(i int) int64 { return be.requests[i].Load() }

func TestEchoBackends_DialerRotates(t *testing.T) {
	be := newBufconnEchoBackends(t)

	for i := 0; i < 4; i++ {
		conn, err := be.dialer()()
		require.NoError(t, err)
		t.Cleanup(func() { _ = conn.Close() })

		_, err = echov1.NewEchoerClient(conn).Echo(context.Background(), &echov1.EchoRequest{})
		require.NoError(t, err)
	}

	assert.Equal(t, int64(4), be.dialCount(), "four ClientConns should drive four context-dialer calls")
	assert.Equal(t, int64(2), be.requestCount(0), "backend 0 should receive half the requests")
	assert.Equal(t, int64(2), be.requestCount(1), "backend 1 should receive half the requests")
}

func multibackendConfig(be *echoBackends, model string, concurrency int) *Config {
	return &Config{
		Target:       "passthrough://bufnet",
		Plaintext:    true,
		Concurrency:  concurrency,
		Duration:     200 * time.Millisecond,
		Payload:      defaultMix,
		Compression:  CompressionIdentity,
		ConnModel:    model,
		OutputFormat: OutputFormatHuman,
		Output:       io.Discard,
		dialOpts:     be.dialOpts(),
	}
}

// inFlightSlack is the maximum number of un-recorded RPCs at run
// termination. A worker can have one call in flight when the duration
// timer fires; that call returns a cancellation error and is dropped by
// doCall, but its server-side interceptor or dial may have already run.
// The bound is one per worker.
func inFlightSlack(cfg *Config) int64 { return int64(cfg.Concurrency) }

func TestRun_PerWorker_SpreadsAcrossBackends(t *testing.T) {
	be := newBufconnEchoBackends(t)
	cfg := multibackendConfig(be, ConnModelPerWorker, 2)

	s, err := cfg.run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, s.ErrorCount)
	require.Greater(t, s.Count, 0)

	assert.Equal(t, int64(2), be.dialCount(),
		"per-worker should dial exactly once per worker")
	assert.Greater(t, be.requestCount(0), int64(0),
		"backend 0 should serve requests from its pinned worker")
	assert.Greater(t, be.requestCount(1), int64(0),
		"backend 1 should serve requests from its pinned worker")

	sum := be.requestCount(0) + be.requestCount(1)
	assert.GreaterOrEqual(t, sum, int64(s.Count),
		"every recorded RPC was served by one of the backends")
	assert.LessOrEqual(t, sum-int64(s.Count), inFlightSlack(cfg),
		"un-recorded calls that reached the server bounded by one per worker")
}

func TestRun_Shared_PinsToOneBackend(t *testing.T) {
	be := newBufconnEchoBackends(t)
	cfg := multibackendConfig(be, ConnModelShared, 2)

	s, err := cfg.run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, s.ErrorCount)
	require.Greater(t, s.Count, 0)

	assert.Equal(t, int64(1), be.dialCount(),
		"shared should dial exactly once for the whole run")
	assert.Equal(t, int64(0), be.requestCount(1),
		"backend 1 should be untouched while the shared conn is pinned elsewhere")

	assert.GreaterOrEqual(t, be.requestCount(0), int64(s.Count),
		"every recorded RPC reached the pinned backend (kube-proxy L4 model)")
	assert.LessOrEqual(t, be.requestCount(0)-int64(s.Count), inFlightSlack(cfg),
		"un-recorded calls that reached the server bounded by one per worker")
}

func TestRun_PerRequest_DialsPerCall(t *testing.T) {
	be := newBufconnEchoBackends(t)
	cfg := multibackendConfig(be, ConnModelPerRequest, 1)

	s, err := cfg.run(context.Background())
	require.NoError(t, err)
	require.Equal(t, 0, s.ErrorCount)
	require.Greater(t, s.Count, 1,
		"need at least two calls to exercise both rotated backends")

	assert.GreaterOrEqual(t, be.dialCount(), int64(s.Count),
		"per-request should dial at least once per recorded RPC")
	assert.LessOrEqual(t, be.dialCount()-int64(s.Count), inFlightSlack(cfg),
		"the only un-recorded dials are in-flight calls cancelled at termination")

	assert.Greater(t, be.requestCount(0), int64(0),
		"rotating dialer should land some requests on backend 0")
	assert.Greater(t, be.requestCount(1), int64(0),
		"rotating dialer should land some requests on backend 1")

	sum := be.requestCount(0) + be.requestCount(1)
	assert.GreaterOrEqual(t, sum, int64(s.Count),
		"every recorded RPC was served by one of the backends")
	assert.LessOrEqual(t, sum, be.dialCount(),
		"no backend served more RPCs than were dialed")
}
