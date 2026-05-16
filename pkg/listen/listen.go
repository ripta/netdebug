package listen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"k8s.io/klog/v2"
)

type Server struct {
	Host    string
	Port    int
	Network string

	listenerMu sync.RWMutex
	listener   net.Listener
}

func New() *Server {
	return &Server{
		Host:    "127.0.0.1",
		Port:    0, // random
		Network: "tcp",
	}
}

// Addr returns the bound listener address, or nil if Run has not bound a
// listener yet. Safe to call concurrently with Run.
func (s *Server) Addr() net.Addr {
	s.listenerMu.RLock()
	defer s.listenerMu.RUnlock()
	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

func (s *Server) Run(ctx context.Context) error {
	// TODO(ripta): handle non-TCP case
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	l, err := net.Listen(s.Network, addr)
	if err != nil {
		return err
	}

	s.listenerMu.Lock()
	s.listener = l
	s.listenerMu.Unlock()

	defer func() {
		s.listenerMu.Lock()
		s.listener = nil
		s.listenerMu.Unlock()
	}()

	stopped := make(chan struct{})
	defer close(stopped)
	go func() {
		select {
		case <-ctx.Done():
			if err := l.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
				klog.ErrorS(err, "closing listener")
			}
		case <-stopped:
		}
	}()

	klog.InfoS("listening", "address", l.Addr(), "network", s.Network)
	for {
		c, err := l.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			klog.ErrorS(err, "accepting client connection")
			continue
		}

		go s.handle(c)
	}
}

func (s *Server) handle(c net.Conn) {
	defer func() {
		if err := c.Close(); err != nil {
			klog.ErrorS(err, "closing client connection", "remote_address", c.RemoteAddr())
		}
	}()
	if err := c.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		klog.ErrorS(err, "setting deadline on client connection", "remote_address", c.RemoteAddr())
		return
	}
	klog.InfoS("accepted client connection", "remote_address", c.RemoteAddr(), "local_address", c.LocalAddr())

	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, c); err != nil {
		klog.ErrorS(err, "copying from client connection", "remote_address", c.RemoteAddr())
		return
	}

	klog.InfoS("received payload", "payload", buf.String(), "size_bytes", buf.Len(), "remote_address", c.RemoteAddr())
}
