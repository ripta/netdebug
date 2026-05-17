package bench

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Connection-model selectors. The flag and its semantics mirror NDB-004:
// per-worker reuses one ClientConn per worker, shared multiplexes one
// ClientConn across all workers, and per-request forces a fresh handshake
// per RPC.
const (
	ConnModelPerWorker  = "per-worker"
	ConnModelShared     = "shared"
	ConnModelPerRequest = "per-request"
)

var validConnModels = map[string]struct{}{
	ConnModelPerWorker:  {},
	ConnModelShared:     {},
	ConnModelPerRequest: {},
}

func isValidConnModel(name string) bool {
	_, ok := validConnModels[name]
	return ok
}

// ConnPool owns the long-lived gRPC client resources for a bench run and
// hands out per-worker ConnSources whose behavior depends on the selected
// connection model.
type ConnPool interface {
	NewSource() (ConnSource, error)
	Close() error
}

// ConnSource is the per-worker view of a ConnPool. Each worker calls
// NewSource once at startup and Acquire once per RPC.
type ConnSource interface {
	Acquire(ctx context.Context) (*grpc.ClientConn, ReleaseFunc, error)
	Close() error
}

// ReleaseFunc is returned by Acquire and must be invoked when the caller
// is done with the ClientConn. For per-worker and shared sources it is a
// no-op; for per-request sources it closes the just-dialed connection.
type ReleaseFunc func()

var noopRelease ReleaseFunc = func() {}

type dialFunc func() (*grpc.ClientConn, error)

// newConnPool selects the pool implementation for model. For ConnModelShared
// the underlying connection is dialed eagerly so a startup failure is
// returned to the caller rather than surfacing later as a per-worker error.
func newConnPool(model string, dial dialFunc) (ConnPool, error) {
	switch model {
	case ConnModelPerWorker:
		return &perWorkerPool{dial: dial}, nil
	case ConnModelShared:
		conn, err := dial()
		if err != nil {
			return nil, err
		}
		return &sharedPool{conn: conn}, nil
	case ConnModelPerRequest:
		return &perRequestPool{dial: dial}, nil
	default:
		return nil, fmt.Errorf("unknown conn-model %q", model)
	}
}

type perWorkerPool struct {
	dial dialFunc
}

func (p *perWorkerPool) NewSource() (ConnSource, error) {
	conn, err := p.dial()
	if err != nil {
		return nil, err
	}
	return &perWorkerSource{conn: conn}, nil
}

func (p *perWorkerPool) Close() error { return nil }

type perWorkerSource struct {
	conn *grpc.ClientConn
}

func (s *perWorkerSource) Acquire(_ context.Context) (*grpc.ClientConn, ReleaseFunc, error) {
	return s.conn, noopRelease, nil
}

func (s *perWorkerSource) Close() error {
	if s.conn == nil {
		return nil
	}
	conn := s.conn
	s.conn = nil
	return conn.Close()
}

type sharedPool struct {
	conn *grpc.ClientConn
}

func (p *sharedPool) NewSource() (ConnSource, error) {
	return &sharedSource{conn: p.conn}, nil
}

func (p *sharedPool) Close() error {
	if p.conn == nil {
		return nil
	}
	return p.conn.Close()
}

type sharedSource struct {
	conn *grpc.ClientConn
}

func (s *sharedSource) Acquire(_ context.Context) (*grpc.ClientConn, ReleaseFunc, error) {
	return s.conn, noopRelease, nil
}

func (s *sharedSource) Close() error { return nil }

type perRequestPool struct {
	dial dialFunc
}

func (p *perRequestPool) NewSource() (ConnSource, error) {
	return &perRequestSource{dial: p.dial}, nil
}

func (p *perRequestPool) Close() error { return nil }

type perRequestSource struct {
	dial dialFunc
}

func (s *perRequestSource) Acquire(_ context.Context) (*grpc.ClientConn, ReleaseFunc, error) {
	conn, err := s.dial()
	if err != nil {
		return nil, nil, err
	}
	return conn, func() { _ = conn.Close() }, nil
}

func (s *perRequestSource) Close() error { return nil }

// dial constructs a gRPC client targeting target with the credentials and
// stats handler the bench client always uses. extra dial options are
// appended last so callers can override defaults (e.g., bufconn dialer in
// tests).
func dial(target string, plaintext bool, extra []grpc.DialOption) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if plaintext {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}
	opts := append([]grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithStatsHandler(wireLengthStatsHandler{}),
	}, extra...)
	return grpc.NewClient(target, opts...)
}
