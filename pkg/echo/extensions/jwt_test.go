package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jwtMisconfiguredTest struct {
	Name   string
	Config JWTConfig
}

var jwtMisconfiguredTests = []jwtMisconfiguredTest{
	{
		Name: "missing HeaderName",
		Config: JWTConfig{
			JWKSURL:  "https://example.invalid/jwks.json",
			Audience: "test-audience",
		},
	},
	{
		Name: "missing JWKSURL",
		Config: JWTConfig{
			HeaderName: "X-JWT",
			Audience:   "test-audience",
		},
	},
	{
		Name: "missing Audience",
		Config: JWTConfig{
			HeaderName: "X-JWT",
			JWKSURL:    "https://example.invalid/jwks.json",
		},
	},
}

func TestJWT_ErrMisconfigured(t *testing.T) {
	for _, tc := range jwtMisconfiguredTests {
		t.Run(tc.Name, func(t *testing.T) {
			fn, err := JWT(tc.Config)
			require.ErrorIs(t, err, ErrMisconfigured)
			assert.Nil(t, fn)
		})
	}
}
