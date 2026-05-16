package send

import (
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Run_CopiesReaderPayload(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = l.Close() }()

	received := make(chan []byte, 1)
	go func() {
		c, aerr := l.Accept()
		if aerr != nil {
			received <- nil
			return
		}
		defer func() { _ = c.Close() }()
		data, _ := io.ReadAll(c)
		received <- data
	}()

	cli := &Client{
		Network: "tcp",
		Address: l.Addr().String(),
		Reader:  strings.NewReader("hello-send"),
	}
	require.NoError(t, cli.Run(context.Background()))

	select {
	case got := <-received:
		assert.Equal(t, "hello-send", string(got))
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not receive payload within timeout")
	}
}

func TestClient_Run_PropagatesDialErrors(t *testing.T) {
	// Reserve an address, then close it so the next dial gets refused.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	require.NoError(t, l.Close())

	cli := &Client{
		Network: "tcp",
		Address: addr,
	}
	err = cli.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dialing")
	assert.Contains(t, err.Error(), addr)
}
