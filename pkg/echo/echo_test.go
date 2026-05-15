package echo

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripta/netdebug/pkg/echo/result"
)

// TestExtensionAppearsInResult guards the SA4010 fix: getResultFromRequest
// once built a local extension-result slice that was discarded, while a
// second pass via Extensions.GetResult populated the response. Deleting the
// dead loop must keep the extension visible in the response and must not
// double-invoke the extension.
func TestExtensionAppearsInResult(t *testing.T) {
	calls := 0
	ext := func(_ *http.Request) ([]result.ExtensionResult, error) {
		calls++
		return []result.ExtensionResult{{
			Name: "fake",
			Info: map[string][]string{"k": {"v"}},
		}}, nil
	}

	s := New()
	s.InstallExtension(ext)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	s.echoHandler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var res result.Result
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&res))

	require.Len(t, res.Extensions, 1)
	assert.Equal(t, "fake", res.Extensions[0].Name)
	assert.Equal(t, []string{"v"}, res.Extensions[0].Info["k"])
	assert.Equal(t, 1, calls)
}

func TestErroringExtensionIsSkipped(t *testing.T) {
	ext := func(_ *http.Request) ([]result.ExtensionResult, error) {
		return nil, errors.New("boom")
	}

	s := New()
	s.InstallExtension(ext)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	s.echoHandler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var res result.Result
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&res))

	assert.Empty(t, res.Extensions)
}

func TestHealthzHandler_OK(t *testing.T) {
	s := New()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	s.healthzHandler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK\n", rec.Body.String())
}

func TestEchoHandler_ReflectsRequest_JSON(t *testing.T) {
	s := New()

	req := httptest.NewRequest(http.MethodPost, "/some/path?foo=bar", nil)
	req.Header.Set("X-Test", "value")
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()

	s.echoHandler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var res result.Result
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&res))

	assert.Equal(t, http.MethodPost, res.Request.Method)
	assert.Equal(t, "/some/path?foo=bar", res.Request.URI)
	assert.Equal(t, "/some/path", res.Request.ParsedURL.Path)
	assert.Equal(t, "foo=bar", res.Request.ParsedURL.RawQuery)
	assert.Equal(t, []string{"value"}, res.Request.Headers["X-Test"])
	assert.NotEmpty(t, res.Request.RemoteAddr)
}

func TestEchoHandler_ReflectsRequest_Text(t *testing.T) {
	s := New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.echoHandler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	assert.Contains(t, body, "Request information:")
	assert.Contains(t, body, "Method: GET")
	assert.Contains(t, body, "Raw URI: /")
}

func TestSetupTLSAutogen_LoadsViaTLSConfig(t *testing.T) {
	s := New()

	cleanup, err := s.setupTLSAutogen()
	require.NoError(t, err)
	require.NotNil(t, cleanup)
	defer cleanup()

	require.NotEmpty(t, s.TLSCertPath)
	require.NotEmpty(t, s.TLSKeyPath)

	_, err = os.Stat(s.TLSCertPath)
	require.NoError(t, err)
	_, err = os.Stat(s.TLSKeyPath)
	require.NoError(t, err)

	cert, err := tls.LoadX509KeyPair(s.TLSCertPath, s.TLSKeyPath)
	require.NoError(t, err)
	assert.NotEmpty(t, cert.Certificate)

	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	assert.Len(t, cfg.Certificates, 1)
}
