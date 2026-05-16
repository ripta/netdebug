package bench

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/stats"
)

type compressionValidTest struct {
	Name      string
	Codec     string
	WantValid bool
}

var compressionValidTests = []compressionValidTest{
	{Name: "identity", Codec: CompressionIdentity, WantValid: true},
	{Name: "gzip", Codec: CompressionGzip, WantValid: true},
	{Name: "snappy", Codec: CompressionSnappy, WantValid: true},
	{Name: "zstd", Codec: CompressionZstd, WantValid: true},
	{Name: "empty string is rejected", Codec: "", WantValid: false},
	{Name: "unknown codec is rejected", Codec: "lz4", WantValid: false},
}

func TestIsValidCompression(t *testing.T) {
	for _, tc := range compressionValidTests {
		t.Run(tc.Name, func(t *testing.T) {
			assert.Equal(t, tc.WantValid, isValidCompression(tc.Codec))
		})
	}
}

// Guards against an upstream module rename or codec-name change silently
// dropping a codec from the registry. Identity is special-cased by grpc-go
// and not returned from GetCompressor, so it is not asserted here.
func TestCompressionCodecsRegistered(t *testing.T) {
	for _, name := range []string{CompressionGzip, CompressionSnappy, CompressionZstd} {
		t.Run(name, func(t *testing.T) {
			assert.NotNil(t, encoding.GetCompressor(name), "compressor %q must be registered", name)
		})
	}
}

func TestWireLengthStatsHandlerRecords(t *testing.T) {
	t.Run("records out and in payload sizes onto bag in context", func(t *testing.T) {
		bag := &wireBytes{}
		ctx := contextWithWireBytes(context.Background(), bag)
		h := wireLengthStatsHandler{}

		h.HandleRPC(ctx, &stats.OutPayload{Length: 100, WireLength: 60})
		h.HandleRPC(ctx, &stats.OutPayload{Length: 50, WireLength: 30})
		h.HandleRPC(ctx, &stats.InPayload{Length: 200, WireLength: 120})

		assert.Equal(t, int64(150), bag.SentUncompressed.Load())
		assert.Equal(t, int64(90), bag.SentWire.Load())
		assert.Equal(t, int64(200), bag.ReceivedUncompressed.Load())
		assert.Equal(t, int64(120), bag.ReceivedWire.Load())
	})

	t.Run("ignores events when bag is missing from context", func(t *testing.T) {
		h := wireLengthStatsHandler{}
		assert.NotPanics(t, func() {
			h.HandleRPC(context.Background(), &stats.OutPayload{Length: 10, WireLength: 5})
			h.HandleRPC(context.Background(), &stats.InPayload{Length: 10, WireLength: 5})
		})
	})

	t.Run("ignores unrelated RPCStats events", func(t *testing.T) {
		bag := &wireBytes{}
		ctx := contextWithWireBytes(context.Background(), bag)
		h := wireLengthStatsHandler{}

		h.HandleRPC(ctx, &stats.Begin{})
		h.HandleRPC(ctx, &stats.End{})

		assert.Equal(t, int64(0), bag.SentUncompressed.Load())
		assert.Equal(t, int64(0), bag.SentWire.Load())
		assert.Equal(t, int64(0), bag.ReceivedUncompressed.Load())
		assert.Equal(t, int64(0), bag.ReceivedWire.Load())
	})
}
