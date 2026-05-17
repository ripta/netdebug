package bench

import (
	"context"
	"net"
	"sync/atomic"
	"testing"

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

func newBufconnEchoBackends(t *testing.T, n int) *echoBackends {
	t.Helper()

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
	be := newBufconnEchoBackends(t, 2)

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
