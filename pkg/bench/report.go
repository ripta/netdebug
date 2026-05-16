package bench

import (
	"fmt"
	"io"
	"time"
)

func writeReport(w io.Writer, c *Config, s Summary) error {
	successes := s.Count - s.ErrorCount
	_, err := fmt.Fprintf(w,
		"Target:      %s\n"+
			"Concurrency: %d\n"+
			"Duration:    %s\n"+
			"\n"+
			"Requests:    %d\n"+
			"Errors:      %d\n"+
			"Elapsed:     %s\n"+
			"Throughput:  %.2f req/s\n"+
			"Latency p50: %s\n"+
			"Latency p99: %s\n",
		c.Target,
		c.Concurrency,
		c.Duration,
		s.Count,
		s.ErrorCount,
		s.Elapsed,
		s.Throughput,
		latencyString(successes, s.LatencyP50),
		latencyString(successes, s.LatencyP99),
	)
	return err
}

func latencyString(successes int, d time.Duration) string {
	if successes <= 0 {
		return "n/a"
	}
	return d.String()
}
