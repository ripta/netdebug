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

type buildEchoRequestTest struct {
	Name   string
	Shape  echov1.PayloadShape
	Sizes  PayloadSizes
	Verify func(t *testing.T, req *echov1.EchoRequest)
}

var buildEchoRequestTests = []buildEchoRequestTest{
	{
		Name:  "string",
		Shape: echov1.PayloadShape_PAYLOAD_SHAPE_STRING,
		Sizes: PayloadSizes{StringLen: 17},
		Verify: func(t *testing.T, req *echov1.EchoRequest) {
			wrap, ok := req.Payload.(*echov1.EchoRequest_StringPayload)
			require.True(t, ok, "Payload = %T, want *EchoRequest_StringPayload", req.Payload)
			assert.Equal(t, 17, len(wrap.StringPayload))
		},
	},
	{
		Name:  "bytes",
		Shape: echov1.PayloadShape_PAYLOAD_SHAPE_BYTES,
		Sizes: PayloadSizes{BytesSize: 32},
		Verify: func(t *testing.T, req *echov1.EchoRequest) {
			wrap, ok := req.Payload.(*echov1.EchoRequest_BytesPayload)
			require.True(t, ok, "Payload = %T, want *EchoRequest_BytesPayload", req.Payload)
			assert.Equal(t, 32, len(wrap.BytesPayload))
		},
	},
	{
		Name:  "embedding-float",
		Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT,
		Sizes: PayloadSizes{EmbeddingDim: 256},
		Verify: func(t *testing.T, req *echov1.EchoRequest) {
			wrap, ok := req.Payload.(*echov1.EchoRequest_EmbeddingFloat)
			require.True(t, ok, "Payload = %T, want *EchoRequest_EmbeddingFloat", req.Payload)
			require.NotNil(t, wrap.EmbeddingFloat)
			assert.Equal(t, 256, len(wrap.EmbeddingFloat.Values))
		},
	},
	{
		Name:  "embedding-bytes (4 bytes per dim)",
		Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES,
		Sizes: PayloadSizes{EmbeddingDim: 256},
		Verify: func(t *testing.T, req *echov1.EchoRequest) {
			wrap, ok := req.Payload.(*echov1.EchoRequest_EmbeddingBytes)
			require.True(t, ok, "Payload = %T, want *EchoRequest_EmbeddingBytes", req.Payload)
			assert.Equal(t, 1024, len(wrap.EmbeddingBytes))
		},
	},
	{
		Name:  "mixed populates Name and Blob from string/bytes sizes",
		Shape: echov1.PayloadShape_PAYLOAD_SHAPE_MIXED,
		Sizes: PayloadSizes{StringLen: 9, BytesSize: 13},
		Verify: func(t *testing.T, req *echov1.EchoRequest) {
			wrap, ok := req.Payload.(*echov1.EchoRequest_Mixed)
			require.True(t, ok, "Payload = %T, want *EchoRequest_Mixed", req.Payload)
			require.NotNil(t, wrap.Mixed)
			assert.Equal(t, 9, len(wrap.Mixed.Name))
			assert.Equal(t, 13, len(wrap.Mixed.Blob))
		},
	},
	{
		Name:  "zero sizes produce empty payloads",
		Shape: echov1.PayloadShape_PAYLOAD_SHAPE_BYTES,
		Sizes: PayloadSizes{},
		Verify: func(t *testing.T, req *echov1.EchoRequest) {
			wrap, ok := req.Payload.(*echov1.EchoRequest_BytesPayload)
			require.True(t, ok, "Payload = %T, want *EchoRequest_BytesPayload", req.Payload)
			assert.Equal(t, 0, len(wrap.BytesPayload))
		},
	},
}

func TestBuildEchoRequest(t *testing.T) {
	for _, tc := range buildEchoRequestTests {
		t.Run(tc.Name, func(t *testing.T) {
			req := BuildEchoRequest(tc.Shape, tc.Sizes)
			require.NotNil(t, req)
			assert.Equal(t, tc.Shape, req.Shape)
			tc.Verify(t, req)
		})
	}
}

func TestBuildEchoRequest_UnspecifiedShapeHasNoPayload(t *testing.T) {
	req := BuildEchoRequest(echov1.PayloadShape_PAYLOAD_SHAPE_UNSPECIFIED, PayloadSizes{
		EmbeddingDim: 16, BytesSize: 16, StringLen: 16,
	})
	require.NotNil(t, req)
	assert.Equal(t, echov1.PayloadShape_PAYLOAD_SHAPE_UNSPECIFIED, req.Shape)
	assert.Nil(t, req.Payload)
}

func TestPayloadMix_Set_Success(t *testing.T) {
	var m PayloadMix
	require.NoError(t, m.Set("embedding-float:50,embedding-bytes:50"))
	assert.Equal(t, PayloadMix{
		{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 50},
		{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 50},
	}, m)
}

func TestPayloadMix_Set_FailureLeavesReceiverUnchanged(t *testing.T) {
	m := PayloadMix{{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_STRING, Weight: 1}}
	original := append(PayloadMix(nil), m...)
	err := m.Set("garbage")
	require.Error(t, err)
	assert.Equal(t, original, m)
}

func TestPayloadMix_String_RoundTrip(t *testing.T) {
	cases := []PayloadMix{
		{{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1}},
		{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 50},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 50},
		},
		{
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_STRING, Weight: 3},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_BYTES, Weight: 7},
			{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_MIXED, Weight: 1},
		},
	}
	for _, want := range cases {
		t.Run(want.String(), func(t *testing.T) {
			got, err := ParsePayloadMix(want.String())
			require.NoError(t, err)
			assert.Equal(t, want, got)
		})
	}
}

func TestPayloadMix_String_Empty(t *testing.T) {
	var m PayloadMix
	assert.Equal(t, "", m.String())
}

func TestPayloadMix_Type(t *testing.T) {
	var m PayloadMix
	assert.Equal(t, "payload-mix", m.Type())
}
