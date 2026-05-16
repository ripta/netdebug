package bench

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

type parsePayloadMixValidTest struct {
	Name  string
	Input string
	Want  PayloadMix
}

var parsePayloadMixValidTests = []parsePayloadMixValidTest{
	{
		Name:  "single string",
		Input: "string",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_STRING, Weight: 1},
		},
	},
	{
		Name:  "single embedding-float",
		Input: "embedding-float",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1},
		},
	},
	{
		Name:  "default-equal weights",
		Input: "embedding-float,embedding-bytes",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 1},
		},
	},
	{
		Name:  "equal explicit weights",
		Input: "embedding-float:50,embedding-bytes:50",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 50},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 50},
		},
	},
	{
		Name:  "unequal explicit weights",
		Input: "embedding-float:30,embedding-bytes:70",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 30},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 70},
		},
	},
	{
		Name:  "zero weight accepted",
		Input: "embedding-float:0,embedding-bytes:50",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 0},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 50},
		},
	},
	{
		Name:  "all five shapes",
		Input: "string:1,bytes:1,embedding-float:1,embedding-bytes:1,mixed:1",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_STRING, Weight: 1},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_BYTES, Weight: 1},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 1},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_MIXED, Weight: 1},
		},
	},
	{
		Name:  "surrounding whitespace trimmed",
		Input: "  embedding-float  ",
		Want: PayloadMix{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1},
		},
	},
}

func TestParsePayloadMix_Valid(t *testing.T) {
	for _, tc := range parsePayloadMixValidTests {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := ParsePayloadMix(tc.Input)
			require.NoError(t, err)
			assert.Equal(t, tc.Want, got)
		})
	}
}

type parsePayloadMixInvalidTest struct {
	Name    string
	Input   string
	WantErr error
}

var parsePayloadMixInvalidTests = []parsePayloadMixInvalidTest{
	{Name: "empty string", Input: "", WantErr: ErrEmptyMix},
	{Name: "only whitespace", Input: "   ", WantErr: ErrEmptyMix},
	{Name: "unknown shape", Input: "floats", WantErr: ErrUnknownShape},
	{Name: "mixed casing rejected", Input: "Embedding-Float", WantErr: ErrUnknownShape},
	{Name: "negative weight", Input: "embedding-float:-1", WantErr: ErrInvalidWeight},
	{Name: "non-integer weight", Input: "embedding-float:abc", WantErr: ErrInvalidWeight},
	{Name: "missing weight after colon", Input: "embedding-float:", WantErr: ErrInvalidWeight},
	{Name: "missing shape before colon", Input: ":50", WantErr: ErrMalformedToken},
	{Name: "empty token between commas", Input: "embedding-float,,embedding-bytes", WantErr: ErrMalformedToken},
	{Name: "trailing comma", Input: "embedding-float,", WantErr: ErrMalformedToken},
	{Name: "leading comma", Input: ",embedding-float", WantErr: ErrMalformedToken},
	{Name: "extra colon", Input: "embedding-float:50:60", WantErr: ErrMalformedToken},
	{Name: "duplicate shape", Input: "embedding-float:30,embedding-float:20", WantErr: ErrDuplicateShape},
	{Name: "mixed weighting", Input: "embedding-float,embedding-bytes:50", WantErr: ErrMixedWeighting},
	{Name: "mixed weighting reversed", Input: "embedding-float:50,embedding-bytes", WantErr: ErrMixedWeighting},
}

func TestParsePayloadMix_Invalid(t *testing.T) {
	for _, tc := range parsePayloadMixInvalidTests {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := ParsePayloadMix(tc.Input)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.True(t, errors.Is(err, tc.WantErr), "err = %v, want errors.Is(_, %v)", err, tc.WantErr)
		})
	}
}
