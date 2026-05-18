package bench

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc/codes"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

// jsonReport is the on-wire JSON schema. It mirrors the human reporter's
// layout: a config block describing the run knobs and a summary block
// describing the observed results. Schema-stability and round-trip tests
// live in their own milestone; this file owns the schema.
type jsonReport struct {
	Config  jsonConfig  `json:"config"`
	Summary jsonSummary `json:"summary"`
}

type jsonConfig struct {
	Target       string             `json:"target"`
	Plaintext    bool               `json:"plaintext"`
	Concurrency  int                `json:"concurrency"`
	Duration     jsonDuration       `json:"duration"`
	Payload      []jsonPayloadEntry `json:"payload"`
	EmbeddingDim int                `json:"embedding_dim"`
	BytesSize    int                `json:"bytes_size"`
	StringLen    int                `json:"string_len"`
	Compression  string             `json:"compression"`
	ConnModel    string             `json:"conn_model"`
	Labels       map[string]string  `json:"labels"`
}

type jsonPayloadEntry struct {
	Shape  jsonPayloadShape `json:"shape"`
	Weight int              `json:"weight"`
}

type jsonSummary struct {
	Count       int                `json:"count"`
	ErrorCount  int                `json:"error_count"`
	Elapsed     jsonDuration       `json:"elapsed"`
	Throughput  float64            `json:"throughput_rps"`
	ConnModel   string             `json:"conn_model"`
	Total       jsonLatencyStats   `json:"total"`
	Server      jsonLatencyStats   `json:"server"`
	Network     jsonLatencyStats   `json:"network"`
	Upstream    jsonLatencyStats   `json:"upstream"`
	Backends    []jsonBackendStats `json:"backends"`
	BackendSkew jsonBackendSkew    `json:"backend_skew"`
	Errors      []jsonStatusCode   `json:"errors"`
}

type jsonLatencyStats struct {
	Count  int          `json:"count"`
	Min    jsonDuration `json:"min"`
	Mean   jsonDuration `json:"mean"`
	Stddev jsonDuration `json:"stddev"`
	Max    jsonDuration `json:"max"`
	P50    jsonDuration `json:"p50"`
	P90    jsonDuration `json:"p90"`
	P99    jsonDuration `json:"p99"`
}

type jsonBackendStats struct {
	Key            string       `json:"key"`
	Source         string       `json:"source"`
	Count          int          `json:"count"`
	ErrorCount     int          `json:"error_count"`
	PercentOfTotal float64      `json:"percent_of_total"`
	P50            jsonDuration `json:"p50"`
	P99            jsonDuration `json:"p99"`
}

type jsonBackendSkew struct {
	CountRatio float64 `json:"count_ratio"`
	P99Ratio   float64 `json:"p99_ratio"`
}

type jsonStatusCode struct {
	Code        codes.Code         `json:"code"`
	CodeName    string             `json:"code_name"`
	Count       int                `json:"count"`
	TopMessages []jsonErrorMessage `json:"top_messages"`
}

type jsonErrorMessage struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// jsonDuration marshals a Go time.Duration through its String() form so
// the JSON output reads as "5s" / "3ms" / "500ns" rather than a raw
// nanosecond count. UnmarshalJSON inverts the operation through
// time.ParseDuration, accepting any spelling that ParseDuration accepts;
// the empty string is treated as zero so JSON null and a literal "" both
// round-trip cleanly.
type jsonDuration time.Duration

func (d jsonDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *jsonDuration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("jsonDuration: %w", err)
	}
	if s == "" {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("jsonDuration: parse %q: %w", s, err)
	}
	*d = jsonDuration(parsed)
	return nil
}

// jsonPayloadShape marshals an echov1.PayloadShape through the kebab-case
// flag spelling so JSON consumers see the same identifier they would type
// at --payload. Unknown values fall through to the proto-generated name to
// keep output non-lossy if a shape is added without updating
// payloadShapeFlagNames. UnmarshalJSON accepts both spellings so it can
// invert any string MarshalJSON produces.
type jsonPayloadShape echov1.PayloadShape

func (p jsonPayloadShape) MarshalJSON() ([]byte, error) {
	if name, ok := payloadShapeFlagNames[echov1.PayloadShape(p)]; ok {
		return json.Marshal(name)
	}
	return json.Marshal(echov1.PayloadShape(p).String())
}

func (p *jsonPayloadShape) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("jsonPayloadShape: %w", err)
	}
	if shape, ok := payloadShapeNames[s]; ok {
		*p = jsonPayloadShape(shape)
		return nil
	}
	if shape, ok := echov1.PayloadShape_value[s]; ok {
		*p = jsonPayloadShape(echov1.PayloadShape(shape))
		return nil
	}
	return fmt.Errorf("jsonPayloadShape: unknown shape %q", s)
}

func newJSONReport(c *Config, s Summary) jsonReport {
	return jsonReport{
		Config:  newJSONConfig(c),
		Summary: newJSONSummary(s),
	}
}

func newJSONConfig(c *Config) jsonConfig {
	payload := make([]jsonPayloadEntry, len(c.Payload))
	for i, e := range c.Payload {
		payload[i] = jsonPayloadEntry{
			Shape:  jsonPayloadShape(e.Shape),
			Weight: e.Weight,
		}
	}
	// Force an empty map instead of nil so the JSON renders {} and consumers
	// always see the field with a predictable object type.
	labels := c.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	return jsonConfig{
		Target:       c.Target,
		Plaintext:    c.Plaintext,
		Concurrency:  c.Concurrency,
		Duration:     jsonDuration(c.Duration),
		Payload:      payload,
		EmbeddingDim: c.EmbeddingDim,
		BytesSize:    c.BytesSize,
		StringLen:    c.StringLen,
		Compression:  c.Compression,
		ConnModel:    c.ConnModel,
		Labels:       labels,
	}
}

func newJSONSummary(s Summary) jsonSummary {
	backends := make([]jsonBackendStats, len(s.Backends))
	for i, b := range s.Backends {
		backends[i] = jsonBackendStats{
			Key:            b.Key,
			Source:         b.Source,
			Count:          b.Count,
			ErrorCount:     b.ErrorCount,
			PercentOfTotal: b.PercentOfTotal,
			P50:            jsonDuration(b.P50),
			P99:            jsonDuration(b.P99),
		}
	}
	errs := make([]jsonStatusCode, len(s.Errors))
	for i, e := range s.Errors {
		msgs := make([]jsonErrorMessage, len(e.TopMessages))
		for j, m := range e.TopMessages {
			msgs[j] = jsonErrorMessage(m)
		}
		errs[i] = jsonStatusCode{
			Code:        e.Code,
			CodeName:    e.CodeName,
			Count:       e.Count,
			TopMessages: msgs,
		}
	}
	return jsonSummary{
		Count:       s.Count,
		ErrorCount:  s.ErrorCount,
		Elapsed:     jsonDuration(s.Elapsed),
		Throughput:  s.Throughput,
		ConnModel:   s.ConnModel,
		Total:       toJSONLatencyStats(s.Total),
		Server:      toJSONLatencyStats(s.Server),
		Network:     toJSONLatencyStats(s.Network),
		Upstream:    toJSONLatencyStats(s.Upstream),
		Backends:    backends,
		BackendSkew: jsonBackendSkew{CountRatio: s.BackendSkew.CountRatio, P99Ratio: s.BackendSkew.P99Ratio},
		Errors:      errs,
	}
}

func toJSONLatencyStats(l LatencyStats) jsonLatencyStats {
	return jsonLatencyStats{
		Count:  l.Count,
		Min:    jsonDuration(l.Min),
		Mean:   jsonDuration(l.Mean),
		Stddev: jsonDuration(l.Stddev),
		Max:    jsonDuration(l.Max),
		P50:    jsonDuration(l.P50),
		P90:    jsonDuration(l.P90),
		P99:    jsonDuration(l.P99),
	}
}

func writeJSONReport(w io.Writer, c *Config, s Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(newJSONReport(c, s))
}
