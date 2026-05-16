package v1

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

func newBufconnClient(t *testing.T) EchoerClient {
	t.Helper()

	lis := bufconn.Listen(1 << 20)
	gs := grpc.NewServer()
	RegisterEchoerServer(gs, &Server{})

	go func() {
		if err := gs.Serve(lis); err != nil {
			t.Logf("grpc serve returned: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
		gs.Stop()
	})

	return NewEchoerClient(conn)
}

type roundTripTest struct {
	Name          string
	Request       *EchoRequest
	WantShape     PayloadShape
	AssertPayload func(*testing.T, isEchoResponse_Payload)
}

var roundTripTests = []roundTripTest{
	{
		Name: "string",
		Request: &EchoRequest{
			Shape:   PayloadShape_PAYLOAD_SHAPE_STRING,
			Payload: &EchoRequest_StringPayload{StringPayload: "hello world"},
		},
		WantShape: PayloadShape_PAYLOAD_SHAPE_STRING,
		AssertPayload: func(t *testing.T, p isEchoResponse_Payload) {
			sp, ok := p.(*EchoResponse_StringPayload)
			require.True(t, ok, "payload type = %T, want *EchoResponse_StringPayload", p)
			assert.Equal(t, "hello world", sp.StringPayload)
		},
	},
	{
		Name: "bytes",
		Request: &EchoRequest{
			Shape:   PayloadShape_PAYLOAD_SHAPE_BYTES,
			Payload: &EchoRequest_BytesPayload{BytesPayload: []byte{0xde, 0xad, 0xbe, 0xef}},
		},
		WantShape: PayloadShape_PAYLOAD_SHAPE_BYTES,
		AssertPayload: func(t *testing.T, p isEchoResponse_Payload) {
			bp, ok := p.(*EchoResponse_BytesPayload)
			require.True(t, ok, "payload type = %T, want *EchoResponse_BytesPayload", p)
			assert.Equal(t, []byte{0xde, 0xad, 0xbe, 0xef}, bp.BytesPayload)
		},
	},
	{
		Name: "embedding-float",
		Request: &EchoRequest{
			Shape: PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT,
			Payload: &EchoRequest_EmbeddingFloat{
				EmbeddingFloat: &EmbeddingFloat{
					Values: []float32{0, 1, -1, 0.5, -0.5, 3.14, 2.718, 1e-9},
				},
			},
		},
		WantShape: PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT,
		AssertPayload: func(t *testing.T, p isEchoResponse_Payload) {
			ef, ok := p.(*EchoResponse_EmbeddingFloat)
			require.True(t, ok, "payload type = %T, want *EchoResponse_EmbeddingFloat", p)
			require.NotNil(t, ef.EmbeddingFloat)
			assert.Equal(t, []float32{0, 1, -1, 0.5, -0.5, 3.14, 2.718, 1e-9}, ef.EmbeddingFloat.Values)
		},
	},
	{
		Name: "embedding-bytes",
		Request: &EchoRequest{
			Shape: PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES,
			Payload: &EchoRequest_EmbeddingBytes{
				EmbeddingBytes: []byte{
					0x00, 0x00, 0x80, 0x3f,
					0x00, 0x00, 0x00, 0x40,
					0x00, 0x00, 0x40, 0x40,
					0x00, 0x00, 0x80, 0x40,
				},
			},
		},
		WantShape: PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES,
		AssertPayload: func(t *testing.T, p isEchoResponse_Payload) {
			eb, ok := p.(*EchoResponse_EmbeddingBytes)
			require.True(t, ok, "payload type = %T, want *EchoResponse_EmbeddingBytes", p)
			assert.Equal(t, []byte{
				0x00, 0x00, 0x80, 0x3f,
				0x00, 0x00, 0x00, 0x40,
				0x00, 0x00, 0x40, 0x40,
				0x00, 0x00, 0x80, 0x40,
			}, eb.EmbeddingBytes)
		},
	},
	{
		Name: "mixed",
		Request: &EchoRequest{
			Shape: PayloadShape_PAYLOAD_SHAPE_MIXED,
			Payload: &EchoRequest_Mixed{
				Mixed: &MixedPayload{
					Id:      42,
					Name:    "mix",
					Count:   7,
					Ratio:   3.14,
					Enabled: true,
					Tags:    []string{"a", "b", "c"},
					Numbers: []int32{1, 2, 3},
					Attrs:   map[string]string{"k1": "v1", "k2": "v2"},
					Blob:    []byte{0xde, 0xad, 0xbe, 0xef},
					Extra:   &KeyMultivalue{Key: "extra", Values: []string{"e1", "e2"}},
				},
			},
		},
		WantShape: PayloadShape_PAYLOAD_SHAPE_MIXED,
		AssertPayload: func(t *testing.T, p isEchoResponse_Payload) {
			mp, ok := p.(*EchoResponse_Mixed)
			require.True(t, ok, "payload type = %T, want *EchoResponse_Mixed", p)
			want := &MixedPayload{
				Id:      42,
				Name:    "mix",
				Count:   7,
				Ratio:   3.14,
				Enabled: true,
				Tags:    []string{"a", "b", "c"},
				Numbers: []int32{1, 2, 3},
				Attrs:   map[string]string{"k1": "v1", "k2": "v2"},
				Blob:    []byte{0xde, 0xad, 0xbe, 0xef},
				Extra:   &KeyMultivalue{Key: "extra", Values: []string{"e1", "e2"}},
			}
			assert.True(t, proto.Equal(want, mp.Mixed), "mixed payload diff:\nwant: %+v\ngot:  %+v", want, mp.Mixed)
		},
	},
}

func TestEcho_RoundTrip(t *testing.T) {
	client := newBufconnClient(t)
	for _, tc := range roundTripTests {
		t.Run(tc.Name, func(t *testing.T) {
			rsp, err := client.Echo(context.Background(), tc.Request)
			require.NoError(t, err)
			require.NotNil(t, rsp)

			assert.Equal(t, tc.WantShape, rsp.Shape)
			tc.AssertPayload(t, rsp.Payload)

			assert.LessOrEqual(t, rsp.ReceivedAtUnixNano, rsp.SentAtUnixNano)
			assert.GreaterOrEqual(t, rsp.ServerDurationNs, int64(0))
		})
	}
}

func TestEcho_RoundTrip_TimingBounds(t *testing.T) {
	client := newBufconnClient(t)

	before := time.Now().UnixNano()
	rsp, err := client.Echo(context.Background(), &EchoRequest{})
	after := time.Now().UnixNano()
	require.NoError(t, err)
	require.NotNil(t, rsp)

	assert.GreaterOrEqual(t, rsp.ReceivedAtUnixNano, before)
	assert.LessOrEqual(t, rsp.SentAtUnixNano, after)
	assert.LessOrEqual(t, rsp.ReceivedAtUnixNano, rsp.SentAtUnixNano)
	assert.GreaterOrEqual(t, rsp.ServerDurationNs, int64(0))
	assert.LessOrEqual(t, rsp.ServerDurationNs, after-before)
}
