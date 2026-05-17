package bench

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

// OutputFormat values accepted by --output. The human form is the default
// and matches the existing stdout layout; the json form emits a single
// pretty-printed object.
const (
	OutputFormatHuman = "human"
	OutputFormatJSON  = "json"
)

type Config struct {
	Target       string
	Plaintext    bool
	Concurrency  int
	Duration     time.Duration
	Payload      PayloadMix
	EmbeddingDim int
	BytesSize    int
	StringLen    int
	Compression  string
	ConnModel    string
	OutputFormat string
	Output       io.Writer

	dialOpts []grpc.DialOption
}

func New() *Config {
	return &Config{
		Target:      "127.0.0.1:8080",
		Plaintext:   true,
		Concurrency: 1,
		Duration:    10 * time.Second,
		Payload: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1},
		},
		EmbeddingDim: 1024,
		BytesSize:    1024,
		StringLen:    1024,
		Compression:  CompressionIdentity,
		ConnModel:    ConnModelPerWorker,
		OutputFormat: OutputFormatHuman,
		Output:       os.Stdout,
	}
}

func isValidOutputFormat(s string) bool {
	switch s {
	case OutputFormatHuman, OutputFormatJSON:
		return true
	}
	return false
}

func (c *Config) Validate() error {
	if c.Target == "" {
		return errors.New("target must not be empty")
	}
	if c.Concurrency < 1 {
		return errors.New("concurrency must be at least 1")
	}
	if c.Duration <= 0 {
		return errors.New("duration must be greater than zero")
	}
	if len(c.Payload) == 0 {
		return errors.New("payload mix must not be empty")
	}
	hasPositive := false
	for _, e := range c.Payload {
		if e.Weight > 0 {
			hasPositive = true
			break
		}
	}
	if !hasPositive {
		return errors.New("payload mix must contain at least one entry with weight > 0")
	}
	if c.EmbeddingDim < 0 {
		return errors.New("embedding-dim must be >= 0")
	}
	if c.BytesSize < 0 {
		return errors.New("bytes-size must be >= 0")
	}
	if c.StringLen < 0 {
		return errors.New("string-len must be >= 0")
	}
	if !isValidCompression(c.Compression) {
		return fmt.Errorf("compression %q is not one of identity, gzip, snappy, zstd", c.Compression)
	}
	if !isValidConnModel(c.ConnModel) {
		return fmt.Errorf("conn-model %q is not one of per-worker, shared, per-request", c.ConnModel)
	}
	if !isValidOutputFormat(c.OutputFormat) {
		return fmt.Errorf("output %q is not one of human, json", c.OutputFormat)
	}
	return nil
}

func (c *Config) Run(ctx context.Context) error {
	s, err := c.run(ctx)
	if err != nil {
		return err
	}
	switch c.OutputFormat {
	case OutputFormatJSON:
		return writeJSONReport(c.output(), c, s)
	default:
		return writeReport(c.output(), c, s)
	}
}

func (c *Config) run(ctx context.Context) (Summary, error) {
	if err := c.Validate(); err != nil {
		return Summary{}, err
	}

	pool, err := newConnPool(c.ConnModel, func() (*grpc.ClientConn, error) {
		return dial(c.Target, c.Plaintext, c.dialOpts)
	})
	if err != nil {
		return Summary{}, fmt.Errorf("creating conn pool: %w", err)
	}
	defer func() {
		if cerr := pool.Close(); cerr != nil {
			klog.ErrorS(cerr, "closing bench conn pool", "target", c.Target)
		}
	}()

	runCtx, cancel := context.WithTimeout(ctx, c.Duration)
	defer cancel()

	selector := NewPayloadSelector(c.Payload)
	sizes := PayloadSizes{
		EmbeddingDim: c.EmbeddingDim,
		BytesSize:    c.BytesSize,
		StringLen:    c.StringLen,
	}

	workers := make([]*worker, c.Concurrency)
	for i := range workers {
		workers[i] = &worker{
			pool:        pool,
			compression: c.Compression,
			selector:    selector,
			sizes:       sizes,
			rng:         rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		}
	}

	var wg sync.WaitGroup
	for _, w := range workers {
		wg.Add(1)
		go func(w *worker) {
			defer wg.Done()
			w.run(runCtx)
		}(w)
	}
	wg.Wait()

	total := 0
	for _, w := range workers {
		total += len(w.results)
	}
	results := make([]Result, 0, total)
	for _, w := range workers {
		results = append(results, w.results...)
	}

	return aggregate(results, elapsed(results), c.ConnModel), nil
}

func (c *Config) output() io.Writer {
	if c.Output == nil {
		return os.Stdout
	}
	return c.Output
}

func elapsed(results []Result) time.Duration {
	if len(results) == 0 {
		return 0
	}
	earliest := results[0].Start
	latest := results[0].End
	for _, r := range results[1:] {
		if r.Start.Before(earliest) {
			earliest = r.Start
		}
		if r.End.After(latest) {
			latest = r.End
		}
	}
	return latest.Sub(earliest)
}
