package bench

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc/stats"

	// Blank imports register the named codecs in grpc-go's global codec
	// registry. Identity is always available; gzip ships with grpc-go; snappy
	// and zstd come from a community encoder package. The nonclobbering
	// variants skip registration if a codec by the same name is already
	// installed, so importing pkg/bench from another module will not silently
	// overwrite a codec the host has set up.
	_ "github.com/mostynb/go-grpc-compression/nonclobbering/snappy"
	_ "github.com/mostynb/go-grpc-compression/nonclobbering/zstd"
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	CompressionIdentity = "identity"
	CompressionGzip     = "gzip"
	CompressionSnappy   = "snappy"
	CompressionZstd     = "zstd"
)

var validCompressions = map[string]struct{}{
	CompressionIdentity: {},
	CompressionGzip:     {},
	CompressionSnappy:   {},
	CompressionZstd:     {},
}

func isValidCompression(name string) bool {
	_, ok := validCompressions[name]
	return ok
}

// wireBytes accumulates per-RPC byte counts captured by
// wireLengthStatsHandler. Atomic because grpc-go may dispatch payload
// stats from a goroutine other than the caller's.
type wireBytes struct {
	SentUncompressed     atomic.Int64
	SentWire             atomic.Int64
	ReceivedUncompressed atomic.Int64
	ReceivedWire         atomic.Int64
}

type wireBytesKey struct{}

func contextWithWireBytes(ctx context.Context, b *wireBytes) context.Context {
	return context.WithValue(ctx, wireBytesKey{}, b)
}

func wireBytesFromContext(ctx context.Context) *wireBytes {
	b, _ := ctx.Value(wireBytesKey{}).(*wireBytes)
	return b
}

// wireLengthStatsHandler implements grpc/stats.Handler to record per-RPC
// payload sizes. Length is the marshaled message size; WireLength is the
// compressed payload plus gRPC framing. The caller attaches a wireBytes
// pointer to the per-call context; this handler writes to it on
// OutPayload / InPayload events.
type wireLengthStatsHandler struct{}

func (wireLengthStatsHandler) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

func (wireLengthStatsHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	b := wireBytesFromContext(ctx)
	if b == nil {
		return
	}
	switch p := s.(type) {
	case *stats.OutPayload:
		b.SentUncompressed.Add(int64(p.Length))
		b.SentWire.Add(int64(p.WireLength))
	case *stats.InPayload:
		b.ReceivedUncompressed.Add(int64(p.Length))
		b.ReceivedWire.Add(int64(p.WireLength))
	}
}

func (wireLengthStatsHandler) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (wireLengthStatsHandler) HandleConn(context.Context, stats.ConnStats) {}
