package bench

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

// headerEnvoyUpstreamTime is the response header Envoy adds when an
// Istio sidecar sits in front of the backend. Its value is the integer
// number of milliseconds the upstream took, as observed by the sidecar.
// Linkerd2-proxy does not emit an equivalent header.
const headerEnvoyUpstreamTime = "x-envoy-upstream-service-time"

type worker struct {
	pool        ConnPool
	compression string
	selector    *PayloadSelector
	sizes       PayloadSizes
	headerKV    []string
	rng         *rand.Rand
	results     []Result
}

func (w *worker) run(ctx context.Context) {
	src, err := w.pool.NewSource()
	if err != nil {
		now := time.Now()
		w.results = append(w.results, Result{
			Start: now,
			End:   now,
			Err:   fmt.Errorf("opening conn source: %w", err),
		})
		return
	}
	defer func() {
		if cerr := src.Close(); cerr != nil {
			klog.ErrorS(cerr, "closing bench conn source")
		}
	}()

	for ctx.Err() == nil {
		conn, release, err := src.Acquire(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			now := time.Now()
			w.results = append(w.results, Result{
				Start: now,
				End:   now,
				Err:   fmt.Errorf("acquiring conn: %w", err),
			})
			continue
		}
		w.doCall(ctx, conn)
		release()
	}
}

func (w *worker) doCall(ctx context.Context, conn *grpc.ClientConn) {
	client := echov1.NewEchoerClient(conn)
	req := BuildEchoRequest(w.selector.Pick(w.rng), w.sizes)
	bag := &wireBytes{}
	callCtx := contextWithWireBytes(ctx, bag)
	if len(w.headerKV) > 0 {
		callCtx = metadata.AppendToOutgoingContext(callCtx, w.headerKV...)
	}
	var peerInfo peer.Peer
	var hdrMD metadata.MD
	start := time.Now()
	rsp, err := client.Echo(callCtx, req,
		grpc.UseCompressor(w.compression),
		grpc.Peer(&peerInfo),
		grpc.Header(&hdrMD),
	)
	end := time.Now()

	if err != nil && (ctx.Err() != nil || isCancellation(err)) {
		return
	}

	r := Result{
		Start:                     start,
		End:                       end,
		TotalDuration:             end.Sub(start),
		BytesSentUncompressed:     bag.SentUncompressed.Load(),
		BytesSentWire:             bag.SentWire.Load(),
		BytesReceivedUncompressed: bag.ReceivedUncompressed.Load(),
		BytesReceivedWire:         bag.ReceivedWire.Load(),
		Err:                       err,
	}
	if peerInfo.Addr != nil {
		r.PeerAddr = peerInfo.Addr.String()
	}
	if err == nil && rsp != nil {
		r.ServerDurationNs = rsp.ServerDurationNs
		if rsp.Kubernetes != nil {
			r.PodName = rsp.Kubernetes.PodName
			r.PodHostname = rsp.Kubernetes.Hostname
		}
		if vals := hdrMD.Get(headerEnvoyUpstreamTime); len(vals) > 0 {
			if ms, perr := strconv.ParseInt(vals[0], 10, 64); perr == nil {
				r.HasUpstreamTime = true
				r.UpstreamDurationNs = ms * int64(time.Millisecond)
			}
		}
	}
	w.results = append(w.results, r)
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
