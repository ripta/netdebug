package bench

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/stats"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

// grpcFrameOverhead is the per-message gRPC frame header: 1 byte compressed
// flag + 4 byte big-endian length. stats.OutPayload.WireLength and
// stats.InPayload.WireLength include this header in addition to the
// (possibly compressed) payload.
const grpcFrameOverhead = 5

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

type roundTripCodecTest struct {
	Name  string
	Codec string
}

var roundTripCodecTests = []roundTripCodecTest{
	{Name: "identity", Codec: CompressionIdentity},
	{Name: "gzip", Codec: CompressionGzip},
	{Name: "snappy", Codec: CompressionSnappy},
	{Name: "zstd", Codec: CompressionZstd},
}

// TestRoundTripPerCodec exercises each registered codec end-to-end against
// a bufconn echo server. Each subtest sends a zero-filled embedding-bytes
// payload, asserts the response round-trips bit-for-bit, and asserts the
// wireLengthStatsHandler recorded plausible byte counts for that codec.
//
// Identity is asserted to produce wire-bytes equal to uncompressed-bytes
// plus the 5-byte gRPC frame header. gzip / snappy / zstd are asserted to
// produce wire-bytes strictly smaller than uncompressed-bytes; with a 4096
// byte run of zeros each compressed codec shrinks the payload dramatically,
// so the inequality is safe and a regression to identity is caught.
func TestRoundTripPerCodec(t *testing.T) {
	for _, tc := range roundTripCodecTests {
		t.Run(tc.Name, func(t *testing.T) {
			lis := newBufconnEchoServer(t)

			conn, err := dial("passthrough://bufnet", true, bufconnDialOpts(lis))
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })

			client := echov1.NewEchoerClient(conn)

			req := BuildEchoRequest(
				echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES,
				PayloadSizes{EmbeddingDim: 1024},
			)

			bag := &wireBytes{}
			ctx, cancel := context.WithTimeout(
				contextWithWireBytes(context.Background(), bag),
				2*time.Second,
			)
			defer cancel()

			rsp, err := client.Echo(ctx, req, grpc.UseCompressor(tc.Codec))
			require.NoError(t, err)
			require.NotNil(t, rsp)

			assert.Equal(t, req.Shape, rsp.Shape, "response shape must match request")
			assert.Equal(t, req.GetEmbeddingBytes(), rsp.GetEmbeddingBytes(),
				"embedding-bytes payload must round-trip unchanged under %s", tc.Codec)

			sentU := bag.SentUncompressed.Load()
			sentW := bag.SentWire.Load()
			recvU := bag.ReceivedUncompressed.Load()
			recvW := bag.ReceivedWire.Load()

			assert.Positive(t, sentU, "SentUncompressed should be populated")
			assert.Positive(t, sentW, "SentWire should be populated")
			assert.Positive(t, recvU, "ReceivedUncompressed should be populated")
			assert.Positive(t, recvW, "ReceivedWire should be populated")

			if tc.Codec == CompressionIdentity {
				assert.Equal(t, sentU+grpcFrameOverhead, sentW,
					"identity: SentWire must equal SentUncompressed plus the gRPC frame header")
				assert.Equal(t, recvU+grpcFrameOverhead, recvW,
					"identity: ReceivedWire must equal ReceivedUncompressed plus the gRPC frame header")
			} else {
				assert.Less(t, sentW, sentU,
					"%s: SentWire must be smaller than SentUncompressed for a compressible payload", tc.Codec)
				assert.Less(t, recvW, recvU,
					"%s: ReceivedWire must be smaller than ReceivedUncompressed for a compressible payload", tc.Codec)
			}
		})
	}
}
