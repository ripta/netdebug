package bench

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

// Sentinel errors returned by ParsePayloadMix. Tests assert via errors.Is.
var (
	ErrEmptyMix       = errors.New("empty mix")
	ErrUnknownShape   = errors.New("unknown shape")
	ErrInvalidWeight  = errors.New("invalid weight")
	ErrMalformedToken = errors.New("malformed token")
	ErrDuplicateShape = errors.New("duplicate shape")
	ErrMixedWeighting = errors.New("mixed weighted and unweighted entries")
)

// payloadShapeNames maps the flag-facing kebab-case spelling to the proto
// PayloadShape enum. Names are case-sensitive so the grammar surface stays
// tight; help text spells them the same way.
var payloadShapeNames = map[string]echov1.PayloadShape{
	"string":          echov1.PayloadShape_PAYLOAD_SHAPE_STRING,
	"bytes":           echov1.PayloadShape_PAYLOAD_SHAPE_BYTES,
	"embedding-float": echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT,
	"embedding-bytes": echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES,
	"mixed":           echov1.PayloadShape_PAYLOAD_SHAPE_MIXED,
}

// payloadShapeFlagNames is the inverse of payloadShapeNames, used to print
// a PayloadMix back as a flag-shaped string.
var payloadShapeFlagNames = func() map[echov1.PayloadShape]string {
	out := make(map[echov1.PayloadShape]string, len(payloadShapeNames))
	for k, v := range payloadShapeNames {
		out[v] = k
	}
	return out
}()

// PayloadEntry is a single shape in a weighted payload mix.
type PayloadEntry struct {
	Shape  echov1.PayloadShape
	Weight int
}

// PayloadMix is an ordered list of weighted payload shapes parsed from a
// --payload flag value. Entries preserve input order. When the input omits
// weights for every entry, every entry's weight is 1 (default-equal).
type PayloadMix []PayloadEntry

// ParsePayloadMix parses a --payload flag value. Accepted forms:
//
//   - "shape": single shape, weight 1
//   - "shape,shape,...": default-equal weights, weight 1 each
//   - "shape:N,shape:N,...": explicit non-negative integer weights
//
// Either every entry carries a weight or none of them do; mixed forms are
// rejected. Surrounding whitespace is trimmed, but whitespace inside a
// token is not. A weight of zero is accepted and recorded verbatim; the
// consumer decides whether to skip such entries during weighted selection.
func ParsePayloadMix(s string) (PayloadMix, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil, fmt.Errorf("parsing payload mix %q: %w", s, ErrEmptyMix)
	}

	tokens := strings.Split(trimmed, ",")
	mix := make(PayloadMix, 0, len(tokens))
	seen := make(map[echov1.PayloadShape]struct{}, len(tokens))

	weightedCount := 0
	for _, tok := range tokens {
		if tok == "" {
			return nil, fmt.Errorf("parsing payload mix %q: %w", s, ErrMalformedToken)
		}

		parts := strings.Split(tok, ":")
		if len(parts) > 2 {
			return nil, fmt.Errorf("parsing payload mix %q: %w: %q", s, ErrMalformedToken, tok)
		}

		name := parts[0]
		if name == "" {
			return nil, fmt.Errorf("parsing payload mix %q: %w: %q", s, ErrMalformedToken, tok)
		}

		shape, ok := payloadShapeNames[name]
		if !ok {
			return nil, fmt.Errorf("parsing payload mix %q: %w: %q", s, ErrUnknownShape, name)
		}
		if _, dup := seen[shape]; dup {
			return nil, fmt.Errorf("parsing payload mix %q: %w: %q", s, ErrDuplicateShape, name)
		}
		seen[shape] = struct{}{}

		weight := 1
		if len(parts) == 2 {
			weightedCount++
			ws := parts[1]
			if ws == "" {
				return nil, fmt.Errorf("parsing payload mix %q: %w: missing weight after %q", s, ErrInvalidWeight, name)
			}
			w, err := strconv.Atoi(ws)
			if err != nil {
				return nil, fmt.Errorf("parsing payload mix %q: %w: %q", s, ErrInvalidWeight, ws)
			}
			if w < 0 {
				return nil, fmt.Errorf("parsing payload mix %q: %w: %d", s, ErrInvalidWeight, w)
			}
			weight = w
		}

		mix = append(mix, PayloadEntry{Shape: shape, Weight: weight})
	}

	if weightedCount != 0 && weightedCount != len(mix) {
		return nil, fmt.Errorf("parsing payload mix %q: %w", s, ErrMixedWeighting)
	}

	return mix, nil
}

// String renders the mix in a form ParsePayloadMix can re-parse. A
// single-entry mix with weight 1 is rendered as a bare shape name; any
// other mix is rendered as "shape:weight,shape:weight". An empty mix
// renders as the empty string.
func (m *PayloadMix) String() string {
	if m == nil || len(*m) == 0 {
		return ""
	}
	if len(*m) == 1 && (*m)[0].Weight == 1 {
		return payloadShapeFlagNames[(*m)[0].Shape]
	}
	parts := make([]string, 0, len(*m))
	for _, e := range *m {
		parts = append(parts, fmt.Sprintf("%s:%d", payloadShapeFlagNames[e.Shape], e.Weight))
	}
	return strings.Join(parts, ",")
}

// Set parses s through ParsePayloadMix and replaces the receiver on
// success. On error the receiver is left unmodified so cobra/pflag's
// flag-parsing diagnostics report a clean before/after.
func (m *PayloadMix) Set(s string) error {
	parsed, err := ParsePayloadMix(s)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

// Type satisfies pflag.Value and shows up as the placeholder in --help.
func (m *PayloadMix) Type() string {
	return "payload-mix"
}

// PayloadSizes controls how large each payload shape is when built. Units
// differ by shape: EmbeddingDim is the number of float32 dimensions for
// both embedding-float and embedding-bytes, BytesSize is bytes for the
// bytes shape, and StringLen is character count for the string shape.
type PayloadSizes struct {
	EmbeddingDim int
	BytesSize    int
	StringLen    int
}

// PayloadSelector picks shapes from a payload mix according to weight.
// Zero-weight entries are dropped at construction time. A selector is safe
// for concurrent Pick calls as long as each caller supplies its own
// *rand.Rand; the selector itself is read-only after NewPayloadSelector.
type PayloadSelector struct {
	shapes     []echov1.PayloadShape
	cumWeights []int
	total      int
}

// NewPayloadSelector precomputes a weighted picker over the positive-weight
// entries of mix. Returns nil if no entry has weight > 0; callers that have
// already run Config.Validate will not see that case.
func NewPayloadSelector(mix PayloadMix) *PayloadSelector {
	s := &PayloadSelector{
		shapes:     make([]echov1.PayloadShape, 0, len(mix)),
		cumWeights: make([]int, 0, len(mix)),
	}
	for _, e := range mix {
		if e.Weight <= 0 {
			continue
		}
		s.total += e.Weight
		s.shapes = append(s.shapes, e.Shape)
		s.cumWeights = append(s.cumWeights, s.total)
	}
	if len(s.shapes) == 0 {
		return nil
	}
	return s
}

// Pick returns a shape sampled from the configured weight distribution. The
// single-shape mix is a fast path that avoids touching r.
func (s *PayloadSelector) Pick(r *rand.Rand) echov1.PayloadShape {
	if len(s.shapes) == 1 {
		return s.shapes[0]
	}
	n := r.IntN(s.total)
	idx := sort.SearchInts(s.cumWeights, n+1)
	return s.shapes[idx]
}

// BuildEchoRequest constructs an EchoRequest for the given shape, sized
// per the passed-in sizes. Slice and string contents are zero-valued; the
// shape and length are what matter for benchmarking proto-decode and
// wire-size effects. embedding-bytes is sized at 4 * EmbeddingDim so its
// wire size matches embedding-float at the same --embedding-dim. The
// mixed shape places StringLen into Name and BytesSize into Blob; other
// MixedPayload fields are left zero. An unknown or unspecified shape
// returns a request with Shape set and no Payload, so the caller can
// decide whether to treat it as an error.
func BuildEchoRequest(shape echov1.PayloadShape, sizes PayloadSizes) *echov1.EchoRequest {
	req := &echov1.EchoRequest{Shape: shape}
	switch shape {
	case echov1.PayloadShape_PAYLOAD_SHAPE_STRING:
		req.Payload = &echov1.EchoRequest_StringPayload{
			StringPayload: strings.Repeat("x", sizes.StringLen),
		}
	case echov1.PayloadShape_PAYLOAD_SHAPE_BYTES:
		req.Payload = &echov1.EchoRequest_BytesPayload{
			BytesPayload: make([]byte, sizes.BytesSize),
		}
	case echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT:
		req.Payload = &echov1.EchoRequest_EmbeddingFloat{
			EmbeddingFloat: &echov1.EmbeddingFloat{
				Values: make([]float32, sizes.EmbeddingDim),
			},
		}
	case echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES:
		req.Payload = &echov1.EchoRequest_EmbeddingBytes{
			EmbeddingBytes: make([]byte, sizes.EmbeddingDim*4),
		}
	case echov1.PayloadShape_PAYLOAD_SHAPE_MIXED:
		req.Payload = &echov1.EchoRequest_Mixed{
			Mixed: &echov1.MixedPayload{
				Name: strings.Repeat("x", sizes.StringLen),
				Blob: make([]byte, sizes.BytesSize),
			},
		}
	}
	return req
}
