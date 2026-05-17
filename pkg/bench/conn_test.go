package bench

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func bufconnDialer(lis *bufconn.Listener) dialFunc {
	return func() (*grpc.ClientConn, error) {
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		opts = append(opts, bufconnDialOpts(lis)...)
		return grpc.NewClient("passthrough://bufnet", opts...)
	}
}

func TestNewConnPool_UnknownModel(t *testing.T) {
	_, err := newConnPool("round-robin", func() (*grpc.ClientConn, error) {
		return nil, errors.New("should not be called")
	})
	require.Error(t, err)
}

func TestNewConnPool_SharedDialFailurePropagates(t *testing.T) {
	want := errors.New("dial blew up")
	_, err := newConnPool(ConnModelShared, func() (*grpc.ClientConn, error) {
		return nil, want
	})
	require.ErrorIs(t, err, want)
}

func TestPerWorkerPool_SourceReusesConn(t *testing.T) {
	lis := newBufconnEchoServer(t)
	pool, err := newConnPool(ConnModelPerWorker, bufconnDialer(lis))
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	src, err := pool.NewSource()
	require.NoError(t, err)
	t.Cleanup(func() { _ = src.Close() })

	first, release1, err := src.Acquire(context.Background())
	require.NoError(t, err)
	release1()

	second, release2, err := src.Acquire(context.Background())
	require.NoError(t, err)
	release2()

	assert.Same(t, first, second, "per-worker source should reuse its conn across Acquire calls")
}

func TestPerWorkerPool_SourcesAreDistinct(t *testing.T) {
	lis := newBufconnEchoServer(t)
	pool, err := newConnPool(ConnModelPerWorker, bufconnDialer(lis))
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	srcA, err := pool.NewSource()
	require.NoError(t, err)
	t.Cleanup(func() { _ = srcA.Close() })
	srcB, err := pool.NewSource()
	require.NoError(t, err)
	t.Cleanup(func() { _ = srcB.Close() })

	connA, releaseA, err := srcA.Acquire(context.Background())
	require.NoError(t, err)
	releaseA()
	connB, releaseB, err := srcB.Acquire(context.Background())
	require.NoError(t, err)
	releaseB()

	assert.NotSame(t, connA, connB, "per-worker sources should each own their own conn")
}

func TestPerWorkerPool_SourceCloseClosesConn(t *testing.T) {
	lis := newBufconnEchoServer(t)
	pool, err := newConnPool(ConnModelPerWorker, bufconnDialer(lis))
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	src, err := pool.NewSource()
	require.NoError(t, err)
	conn, release, err := src.Acquire(context.Background())
	require.NoError(t, err)
	release()

	require.NoError(t, src.Close())
	assert.Equal(t, connectivity.Shutdown, conn.GetState(),
		"closing the source should shut its conn down")
}

func TestSharedPool_AllSourcesReturnSameConn(t *testing.T) {
	lis := newBufconnEchoServer(t)
	pool, err := newConnPool(ConnModelShared, bufconnDialer(lis))
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	srcA, err := pool.NewSource()
	require.NoError(t, err)
	srcB, err := pool.NewSource()
	require.NoError(t, err)

	connA, releaseA, err := srcA.Acquire(context.Background())
	require.NoError(t, err)
	releaseA()
	connB, releaseB, err := srcB.Acquire(context.Background())
	require.NoError(t, err)
	releaseB()

	assert.Same(t, connA, connB, "shared sources should hand out the pool's single conn")
	// Source.Close is a no-op for shared; the conn should remain usable
	// until Pool.Close runs.
	require.NoError(t, srcA.Close())
	require.NoError(t, srcB.Close())
	assert.NotEqual(t, connectivity.Shutdown, connA.GetState(),
		"shared conn should not be closed by source.Close")
}

func TestSharedPool_CloseShutsDownConn(t *testing.T) {
	lis := newBufconnEchoServer(t)
	pool, err := newConnPool(ConnModelShared, bufconnDialer(lis))
	require.NoError(t, err)

	src, err := pool.NewSource()
	require.NoError(t, err)
	conn, release, err := src.Acquire(context.Background())
	require.NoError(t, err)
	release()

	require.NoError(t, pool.Close())
	assert.Equal(t, connectivity.Shutdown, conn.GetState(),
		"pool.Close should shut the shared conn down")
}

func TestPerRequestPool_AcquireReturnsFreshConnAndReleaseCloses(t *testing.T) {
	lis := newBufconnEchoServer(t)
	pool, err := newConnPool(ConnModelPerRequest, bufconnDialer(lis))
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })

	src, err := pool.NewSource()
	require.NoError(t, err)
	t.Cleanup(func() { _ = src.Close() })

	first, release1, err := src.Acquire(context.Background())
	require.NoError(t, err)
	second, release2, err := src.Acquire(context.Background())
	require.NoError(t, err)

	assert.NotSame(t, first, second, "per-request Acquire should dial a fresh conn each time")

	release1()
	release2()

	require.Eventually(t, func() bool { return first.GetState() == connectivity.Shutdown },
		time.Second, 10*time.Millisecond, "release should shut the just-acquired conn down")
	require.Eventually(t, func() bool { return second.GetState() == connectivity.Shutdown },
		time.Second, 10*time.Millisecond, "release should shut the just-acquired conn down")
}

func TestPerRequestPool_AcquireDialError(t *testing.T) {
	want := errors.New("per-request dial failed")
	pool, err := newConnPool(ConnModelPerRequest, func() (*grpc.ClientConn, error) {
		return nil, want
	})
	require.NoError(t, err)

	src, err := pool.NewSource()
	require.NoError(t, err)

	_, _, err = src.Acquire(context.Background())
	assert.ErrorIs(t, err, want)
}
