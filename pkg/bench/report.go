package bench

import (
	"fmt"
	"io"
	"strings"
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
			"%s\n"+
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
		backendsBlock(s.Backends),
		backendSkewLine(s.BackendSkew),
		statusCodeBlock(s.Errors),
	)
	return err
}

// statusCodeBlock renders the per-status-code error breakdown. Each line
// names the code, its total error count, and up to topErrorMessages
// distinct messages with their counts. Returns "Errors by code: n/a"
// when no errors were observed so a successful run still gets a
// stable-shaped report.
func statusCodeBlock(stats []StatusCodeStats) string {
	if len(stats) == 0 {
		return "Errors by code: n/a"
	}
	var sb strings.Builder
	sb.WriteString("Errors by code:\n")
	for i, st := range stats {
		fmt.Fprintf(&sb, "  %s (%d req): %s", st.CodeName, st.Count, formatTopMessages(st.TopMessages))
		if i < len(stats)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func formatTopMessages(msgs []ErrorMessageStat) string {
	if len(msgs) == 0 {
		return "(no message)"
	}
	parts := make([]string, len(msgs))
	for i, m := range msgs {
		parts[i] = fmt.Sprintf("%q (%d)", m.Message, m.Count)
	}
	return strings.Join(parts, ", ")
}

func backendSkewLine(s BackendSkew) string {
	return fmt.Sprintf("Backend skew: count=%s p99=%s",
		formatRatio(s.CountRatio), formatRatio(s.P99Ratio))
}

func formatRatio(r float64) string {
	if r == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.2f", r)
}

func backendsBlock(backends []BackendStats) string {
	if len(backends) == 0 {
		return "Backends: n/a"
	}
	var sb strings.Builder
	sb.WriteString("Backends:\n")
	for i, b := range backends {
		key := b.Key
		if key == "" {
			key = "(unknown)"
		}
		fmt.Fprintf(&sb, "  %s=%s (%d req, %.1f%%): p50=%s p99=%s errors=%d",
			b.Source, key, b.Count, b.PercentOfTotal, b.P50, b.P99, b.ErrorCount)
		if i < len(backends)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
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
