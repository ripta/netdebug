package echo

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
