package listen

import (
	"bytes"
	"context"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr/funcr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/klog/v2"

	"github.com/ripta/netdebug/pkg/send"
)

// klogCapture collects structured klog output for assertion. handle() logs
// the received payload via klog.InfoS, which is the proposal-approved way to
// verify what the listener actually saw.
type klogCapture struct {
	mu   sync.Mutex
	msgs []string
}

func (c *klogCapture) add(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, s)
}

func (c *klogCapture) contains(s string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, m := range c.msgs {
		if strings.Contains(m, s) {
			return true
		}
	}
	return false
}

func captureKlog(t *testing.T) *klogCapture {
	t.Helper()
	cap := &klogCapture{}
	klog.SetLogger(funcr.New(func(_, args string) {
		cap.add(args)
	}, funcr.Options{}))
	t.Cleanup(klog.ClearLogger)
	return cap
}

func waitContains(t *testing.T, cap *klogCapture, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cap.contains(want) {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("klog did not capture %q within timeout", want)
}

func waitForAddr(t *testing.T, s *Server) net.Addr {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if addr := s.Addr(); addr != nil {
			return addr
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("listener did not bind within timeout")
	return nil
}

func TestServer_Addr_NilBeforeRun(t *testing.T) {
	assert.Nil(t, New().Addr())
}

func TestServer_handle_LogsPayload(t *testing.T) {
	capture := captureKlog(t)

	srv := New()
	srvSide, cliSide := net.Pipe()

	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.handle(srvSide)
	}()

	payload := "hello-handle"
	_, err := cliSide.Write([]byte(payload))
	require.NoError(t, err)
	require.NoError(t, cliSide.Close())

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handle did not return within timeout")
	}

	waitContains(t, capture, payload)
}

func TestServer_Run_RespectsContextCancellation(t *testing.T) {
	captureKlog(t) // silence klog output for cleaner test runs

	srv := New()
	srv.Host = "127.0.0.1"
	srv.Port = 0

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- srv.Run(ctx)
	}()

	addr := waitForAddr(t, srv)
	require.NotNil(t, addr)

	cancel()

	select {
	case err := <-runErr:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}

	assert.Nil(t, srv.Addr())
}

func TestRoundTrip_SendListen(t *testing.T) {
	capture := captureKlog(t)

	srv := New()
	srv.Host = "127.0.0.1"
	srv.Port = 0

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- srv.Run(ctx)
	}()

	addr := waitForAddr(t, srv)

	cli := &send.Client{
		Network: "tcp",
		Address: addr.String(),
		Reader:  bytes.NewReader([]byte("hello-roundtrip")),
	}
	require.NoError(t, cli.Run(context.Background()))

	waitContains(t, capture, "hello-roundtrip")

	cancel()
	select {
	case err := <-runErr:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after ctx cancel")
	}
}
