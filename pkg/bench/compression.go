package bench

import (
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
