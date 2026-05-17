package bench

import (
	"fmt"
	"io"
)

func writeReport(w io.Writer, c *Config, s Summary) error {
	_, err := fmt.Fprintf(w,
		"Target:      %s\n"+
			"Concurrency: %d\n"+
			"Duration:    %s\n"+
			"Conn model:  %s\n"+
			"\n"+
			"Requests:    %d\n"+
			"Errors:      %d\n"+
			"Elapsed:     %s\n"+
			"Throughput:  %.2f req/s\n"+
			"\n"+
			"%s\n"+
			"%s\n"+
			"%s",
		c.Target,
		c.Concurrency,
		c.Duration,
		s.ConnModel,
		s.Count,
		s.ErrorCount,
		s.Elapsed,
		s.Throughput,
		latencyStatsBlock("total", s.Total),
		latencyStatsBlock("server", s.Server),
		latencyStatsBlock("network", s.Network),
	)
	return err
}

func latencyStatsBlock(name string, stats LatencyStats) string {
	if stats.Count == 0 {
		return fmt.Sprintf("Latency (%s): n/a", name)
	}
	return fmt.Sprintf(
		"Latency (%s):\n"+
			"  count:  %d\n"+
			"  min:    %s\n"+
			"  mean:   %s\n"+
			"  stddev: %s\n"+
			"  p50:    %s\n"+
			"  p90:    %s\n"+
			"  p99:    %s\n"+
			"  max:    %s",
		name,
		stats.Count,
		stats.Min,
		stats.Mean,
		stats.Stddev,
		stats.P50,
		stats.P90,
		stats.P99,
		stats.Max,
	)
}
