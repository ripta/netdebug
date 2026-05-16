package bench

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

type worker struct {
	target    string
	plaintext bool
	results   []Result
}

func (w *worker) run(ctx context.Context) {
	conn, err := dial(w.target, w.plaintext)
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

func dial(target string, plaintext bool) (*grpc.ClientConn, error) {
	var creds credentials.TransportCredentials
	if plaintext {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	}
	return grpc.NewClient(target, grpc.WithTransportCredentials(creds))
}
