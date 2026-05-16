package bench

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Config struct {
	Target      string
	Plaintext   bool
	Concurrency int
	Duration    time.Duration

	results []Result
}

func New() *Config {
	return &Config{
		Target:      "127.0.0.1:8080",
		Plaintext:   true,
		Concurrency: 1,
		Duration:    10 * time.Second,
	}
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
	return nil
}

func (c *Config) Run(ctx context.Context) error {
	if err := c.Validate(); err != nil {
		return err
	}

	runCtx, cancel := context.WithTimeout(ctx, c.Duration)
	defer cancel()

	workers := make([]*worker, c.Concurrency)
	for i := range workers {
		workers[i] = &worker{
			target:    c.Target,
			plaintext: c.Plaintext,
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
	c.results = make([]Result, 0, total)
	for _, w := range workers {
		c.results = append(c.results, w.results...)
	}

	return nil
}
