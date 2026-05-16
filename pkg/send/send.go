package send

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"k8s.io/klog/v2"
)

type Client struct {
	Network string
	Address string
	Reader  io.Reader
}

func New() *Client {
	return &Client{
		Network: "tcp",
		Address: "127.0.0.1:8080",
	}
}

func (c *Client) Run(_ context.Context) error {
	conn, err := net.Dial(c.Network, c.Address)
	if err != nil {
		return fmt.Errorf("dialing %s/%s: %w", c.Network, c.Address, err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			klog.ErrorS(err, "closing connection", "remote_address", conn.RemoteAddr())
		}
	}()

	r := c.Reader
	if r == nil {
		r = os.Stdin
	}

	if _, err := io.Copy(conn, r); err != nil {
		return fmt.Errorf("sending payload to %s/%s: %w", c.Network, c.Address, err)
	}

	return nil
}
