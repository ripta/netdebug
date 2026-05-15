package extensions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validJWTConfig() JWTConfig {
	return JWTConfig{
		HeaderName: "X-JWT",
		JWKSURL:    "https://example.invalid/jwks.json",
		Audience:   "test-audience",
	}
}

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

func TestJWT_ConstructsWithRequiredFields(t *testing.T) {
	fn, err := JWT(validJWTConfig())
	require.NoError(t, err)
	require.NotNil(t, fn)
}

func TestJWT_HeaderAbsentReturnsNil(t *testing.T) {
	fn, err := JWT(validJWTConfig())
	require.NoError(t, err)
	require.NotNil(t, fn)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	results, err := fn(req)
	assert.NoError(t, err)
	assert.Nil(t, results)
}
