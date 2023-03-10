package listen

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"k8s.io/klog/v2"
	"net"
	"time"
)

type Server struct {
	Host    string
	Port    int
	Network string
}

func New() *Server {
	return &Server{
		Host:    "127.0.0.1",
		Port:    0, // random
		Network: "tcp",
	}
}

func (s *Server) Run(_ context.Context) error {
	// TODO(ripta): handle non-TCP case
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	l, err := net.Listen(s.Network, addr)
	if err != nil {
		return err
	}

	klog.InfoS("listening", "address", l.Addr(), "network", s.Network)
	for {
		c, err := l.Accept()
		if err != nil {
			klog.ErrorS(err, "accepting client connection")
			continue
		}

		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()
	c.SetDeadline(time.Now().Add(10 * time.Second))
	klog.InfoS("accepted client connection", "remote_address", c.RemoteAddr(), "local_address", c.LocalAddr())

	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, c); err != nil {
		klog.ErrorS(err, "copying from client connection", "remote_address", c.RemoteAddr())
		return
	}

	klog.InfoS("received payload", "payload", buf.String(), "size_bytes", buf.Len(), "remote_address", c.RemoteAddr())
}
