package send

import (
	"io"
	"net"
	"os"
)

type Client struct {
	Network string
	Address string
}

func New() *Client {
	return &Client{
		Network: "tcp",
		Address: "127.0.0.1:8080",
	}
}

func (c *Client) Run() error {
	conn, err := net.Dial(c.Network, c.Address)
	if err != nil {
		return err
	}

	if _, err := io.Copy(conn, os.Stdin); err != nil {
		return err
	}

	return nil
}
