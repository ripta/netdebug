package app

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// docSources are the user-facing markdown files whose `netdebug bench`
// examples are exercised through cobra. Keep the README and every doc that
// shows bench invocations listed here so flag typos surface as test failures.
var docSources = []string{
	"../../README.md",
	"../../docs/bench-local.md",
}

// readBenchExamples scans the configured doc sources for fenced code blocks
// and returns every `netdebug bench` invocation found. Backslash-continuation
// lines are joined; trailing shell redirection (`> file`, `>> file`) is
// stripped so the result tokenizes cleanly into cobra args.
func readBenchExamples(t *testing.T) []string {
	t.Helper()

	var examples []string
	for _, path := range docSources {
		abs, err := filepath.Abs(path)
		require.NoError(t, err)
		examples = append(examples, readBenchExamplesFrom(t, abs)...)
	}
	sort.Strings(examples)
	return examples
}

func readBenchExamplesFrom(t *testing.T, path string) []string {
	t.Helper()

	b, err := os.ReadFile(path)
	require.NoError(t, err, "reading %s", path)

	var (
		examples   []string
		cur        strings.Builder
		collecting bool
		inFence    bool
	)
	flush := func() {
		examples = append(examples, finalizeExample(cur.String()))
		cur.Reset()
		collecting = false
	}
	for _, raw := range strings.Split(string(b), "\n") {
		trim := strings.TrimSpace(raw)
		if strings.HasPrefix(trim, "```") {
			inFence = !inFence
			if collecting {
				flush()
			}
			continue
		}
		if !inFence {
			continue
		}
		if collecting {
			seg, more := stripContinuation(trim)
			cur.WriteByte(' ')
			cur.WriteString(seg)
			if !more {
				flush()
			}
			continue
		}
		if strings.HasPrefix(trim, "netdebug bench") {
			seg, more := stripContinuation(trim)
			cur.WriteString(seg)
			collecting = true
			if !more {
				flush()
			}
		}
	}
	return examples
}

// stripContinuation returns the line minus its trailing backslash and a
// flag indicating whether more continuation lines follow.
func stripContinuation(line string) (string, bool) {
	if strings.HasSuffix(line, "\\") {
		return strings.TrimSpace(strings.TrimSuffix(line, "\\")), true
	}
	return line, false
}

// finalizeExample strips trailing shell redirection (`> file`) since the
// smoke test feeds args to cobra rather than running a real shell. Doc
// examples carry no `>` inside flag values; if that ever changes, this
// will mangle the command and the per-example subtest will surface the
// breakage as an unknown-flag error.
func finalizeExample(s string) string {
	if i := strings.Index(s, ">"); i >= 0 {
		s = s[:i]
	}
	return strings.Join(strings.Fields(s), " ")
}

func TestReadmeBenchExamples_Parse(t *testing.T) {
	examples := readBenchExamples(t)
	require.NotEmpty(t, examples, "no `netdebug bench` examples found in doc sources")

	for _, raw := range examples {
		t.Run(truncateName(raw), func(t *testing.T) {
			tokens := strings.Fields(raw)
			require.GreaterOrEqual(t, len(tokens), 2)
			require.Equal(t, "netdebug", tokens[0])
			require.Equal(t, "bench", tokens[1])

			root := NewRootCommand()
			bench := findSubcommand(root, "bench")
			require.NotNil(t, bench, "bench subcommand missing")
			bench.RunE = func(*cobra.Command, []string) error { return nil }

			root.SetArgs(tokens[1:])
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)

			assert.NoError(t, root.Execute(), "command: %s", raw)
		})
	}
}

func truncateName(s string) string {
	const n = 80
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
