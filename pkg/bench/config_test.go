package bench

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	echov1 "github.com/ripta/netdebug/pkg/echo/v1"
)

func TestNew_Defaults(t *testing.T) {
	c := New()
	require.NotNil(t, c)
	assert.Equal(t, "127.0.0.1:8080", c.Target)
	assert.True(t, c.Plaintext)
	assert.Equal(t, 1, c.Concurrency)
	assert.Equal(t, 10*time.Second, c.Duration)
	assert.Equal(t, PayloadMix{
		{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1},
	}, c.Payload)
	assert.Equal(t, 1024, c.EmbeddingDim)
	assert.Equal(t, 1024, c.BytesSize)
	assert.Equal(t, 1024, c.StringLen)
	assert.NotNil(t, c.Output)
}

func TestNew_DefaultsValidate(t *testing.T) {
	require.NoError(t, New().Validate())
}

// defaultMix is the same payload mix New() installs; existing validation
// rows below want a sane non-empty mix so the new payload checks don't
// trip target/concurrency/duration tests.
var defaultMix = PayloadMix{
	{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 1},
}

type configValidateTest struct {
	Name    string
	Config  Config
	WantErr bool
}

var configValidateTests = []configValidateTest{
	{
		Name: "defaults are valid",
		Config: Config{
			Target: "127.0.0.1:8080", Plaintext: true, Concurrency: 1, Duration: 10 * time.Second,
			Payload: defaultMix, EmbeddingDim: 1024, BytesSize: 1024, StringLen: 1024,
		},
		WantErr: false,
	},
	{
		Name: "empty target is rejected",
		Config: Config{
			Target: "", Concurrency: 1, Duration: time.Second,
			Payload: defaultMix,
		},
		WantErr: true,
	},
	{
		Name: "zero concurrency is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 0, Duration: time.Second,
			Payload: defaultMix,
		},
		WantErr: true,
	},
	{
		Name: "negative concurrency is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: -1, Duration: time.Second,
			Payload: defaultMix,
		},
		WantErr: true,
	},
	{
		Name: "zero duration is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: 0,
			Payload: defaultMix,
		},
		WantErr: true,
	},
	{
		Name: "negative duration is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: -time.Second,
			Payload: defaultMix,
		},
		WantErr: true,
	},
	{
		Name: "empty payload mix is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: time.Second,
			Payload: PayloadMix{},
		},
		WantErr: true,
	},
	{
		Name: "all-zero weights are rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: time.Second,
			Payload: PayloadMix{
				{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_FLOAT, Weight: 0},
				{Shape: echov1.PayloadShape_PAYLOAD_SHAPE_EMBEDDING_BYTES, Weight: 0},
			},
		},
		WantErr: true,
	},
	{
		Name: "negative embedding-dim is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: time.Second,
			Payload: defaultMix, EmbeddingDim: -1,
		},
		WantErr: true,
	},
	{
		Name: "negative bytes-size is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: time.Second,
			Payload: defaultMix, BytesSize: -1,
		},
		WantErr: true,
	},
	{
		Name: "negative string-len is rejected",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: time.Second,
			Payload: defaultMix, StringLen: -1,
		},
		WantErr: true,
	},
	{
		Name: "zero sizes are accepted",
		Config: Config{
			Target: "127.0.0.1:8080", Concurrency: 1, Duration: time.Second,
			Payload: defaultMix, EmbeddingDim: 0, BytesSize: 0, StringLen: 0,
		},
		WantErr: false,
	},
}

func TestConfig_Validate(t *testing.T) {
	for _, tc := range configValidateTests {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Config.Validate()
			if tc.WantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
