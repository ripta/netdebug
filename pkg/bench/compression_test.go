package bench

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/encoding"
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
