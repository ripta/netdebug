package bench

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

type worker struct {
	target    string
	plaintext bool
	dialOpts  []grpc.DialOption
	results   []Result
}

func (w *worker) run(ctx context.Context) {
	conn, err := dial(w.target, w.plaintext, w.dialOpts)
	if err != nil {
		now := time.Now()
		w.results = append(w.results, Result{
			Start: now,
			End:   now,
			Err:   fmt.Errorf("dialing %s: %w", w.target, err),
		})
		return
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			klog.ErrorS(cerr, "closing bench client conn", "target", w.target)
		}
	}()

	client := echov1.NewEchoerClient(conn)

	for ctx.Err() == nil {
		start := time.Now()
		rsp, err := client.Echo(ctx, &echov1.EchoRequest{})
		end := time.Now()

		if err != nil && (ctx.Err() != nil || isCancellation(err)) {
			return
		}

		r := Result{
			Start:         start,
			End:           end,
			TotalDuration: end.Sub(start),
			Err:           err,
		}
		if err == nil && rsp != nil {
			r.ServerDurationNs = rsp.ServerDurationNs
		}
		w.results = append(w.results, r)
	}
}

// isCancellation reports whether an Echo error is a deadline or cancellation
// signal. The bench's own duration timeout shows up either as a parent-context
// error or as a gRPC status with code DeadlineExceeded / Canceled; the latter
// can arrive before the parent context's Done channel is observable. Without
// this check, the tail of every run records a handful of fake errors.
func isCancellation(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	code := status.Code(err)
	return code == codes.Canceled || code == codes.DeadlineExceeded
}

func dial(target string, plaintext bool, extra []grpc.DialOption) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if plaintext {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}
	opts := append([]grpc.DialOption{grpc.WithTransportCredentials(creds)}, extra...)
	return grpc.NewClient(target, opts...)
}
