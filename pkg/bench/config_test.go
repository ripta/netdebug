package bench

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Defaults(t *testing.T) {
	c := New()
	require.NotNil(t, c)
	assert.Equal(t, "127.0.0.1:8080", c.Target)
	assert.True(t, c.Plaintext)
	assert.Equal(t, 1, c.Concurrency)
	assert.Equal(t, 10*time.Second, c.Duration)
}

func TestNew_DefaultsValidate(t *testing.T) {
	require.NoError(t, New().Validate())
}

type configValidateTest struct {
	Name    string
	Config  Config
	WantErr bool
}

var configValidateTests = []configValidateTest{
	{
		Name:    "defaults are valid",
		Config:  Config{Target: "127.0.0.1:8080", Plaintext: true, Concurrency: 1, Duration: 10 * time.Second},
		WantErr: false,
	},
	{
		Name:    "empty target is rejected",
		Config:  Config{Target: "", Concurrency: 1, Duration: time.Second},
		WantErr: true,
	},
	{
		Name:    "zero concurrency is rejected",
		Config:  Config{Target: "127.0.0.1:8080", Concurrency: 0, Duration: time.Second},
		WantErr: true,
	},
	{
		Name:    "negative concurrency is rejected",
		Config:  Config{Target: "127.0.0.1:8080", Concurrency: -1, Duration: time.Second},
		WantErr: true,
	},
	{
		Name:    "zero duration is rejected",
		Config:  Config{Target: "127.0.0.1:8080", Concurrency: 1, Duration: 0},
		WantErr: true,
	},
	{
		Name:    "negative duration is rejected",
		Config:  Config{Target: "127.0.0.1:8080", Concurrency: 1, Duration: -time.Second},
		WantErr: true,
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
