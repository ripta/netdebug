package bench

import (
	"errors"
	"fmt"
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
