package v1

import (
	"context"
	"net/url"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ripta/netdebug/pkg/echo/result"
)

func TestEcho_PopulatesRuntime(t *testing.T) {
	s := &Server{}
	rsp, err := s.Echo(context.Background(), &EchoRequest{Query: "hi"})
	require.NoError(t, err)
	require.NotNil(t, rsp)
	require.NotNil(t, rsp.Runtime)

	assert.Equal(t, "hi", rsp.Query)
	assert.Equal(t, runtime.Version(), rsp.Runtime.GoVersion)
	assert.Equal(t, runtime.GOARCH, rsp.Runtime.GoArch)
	assert.Equal(t, runtime.GOOS, rsp.Runtime.GoOs)
	assert.Equal(t, int64(runtime.NumCPU()), rsp.Runtime.NumCpus)
	assert.Greater(t, rsp.Runtime.NumGoroutines, int64(0))
}

func TestEcho_ReflectsRequest(t *testing.T) {
	res := result.Result{
		Request: result.RequestResult{
			Protocol:   "HTTP/1.1",
			RemoteAddr: "10.0.0.1:5555",
			Method:     "GET",
			URI:        "/echo?x=1&x=2&y=3",
			ParsedURL: result.ParsedURL{
				Scheme:   "http",
				Host:     "example.test",
				Path:     "/echo",
				RawPath:  "/echo",
				RawQuery: "x=1&x=2&y=3",
				Query:    url.Values{"x": {"1", "2"}, "y": {"3"}},
			},
			Headers: map[string][]string{
				"X-Single": {"only"},
				"X-Multi":  {"a", "b"},
			},
		},
	}
	ctx := result.WithResult(context.Background(), res)

	rsp, err := (&Server{}).Echo(ctx, &EchoRequest{})
	require.NoError(t, err)
	require.NotNil(t, rsp)
	require.NotNil(t, rsp.Request)
	require.NotNil(t, rsp.Request.ParsedUrl)

	assert.Equal(t, res.Request.Protocol, rsp.Request.Protocol)
	assert.Equal(t, res.Request.RemoteAddr, rsp.Request.RemoteAddr)
	assert.Equal(t, res.Request.Method, rsp.Request.Method)
	assert.Equal(t, res.Request.URI, rsp.Request.Uri)

	assert.Equal(t, res.Request.ParsedURL.Scheme, rsp.Request.ParsedUrl.Scheme)
	assert.Equal(t, res.Request.ParsedURL.Host, rsp.Request.ParsedUrl.Host)
	assert.Equal(t, res.Request.ParsedURL.Path, rsp.Request.ParsedUrl.Path)
	assert.Equal(t, res.Request.ParsedURL.RawPath, rsp.Request.ParsedUrl.RawPath)
	assert.Equal(t, res.Request.ParsedURL.RawQuery, rsp.Request.ParsedUrl.RawQuery)

	assert.Equal(t, map[string][]string(res.Request.ParsedURL.Query), flatten(rsp.Request.ParsedUrl.Query))
	assert.Equal(t, res.Request.Headers, flatten(rsp.Request.Header))
}

func TestEcho_FlattensExtensions(t *testing.T) {
	res := result.Result{
		Extensions: []result.ExtensionResult{
			{Name: "auth", Info: map[string][]string{
				"subject":  {"alice"},
				"audience": {"a", "b"},
			}},
			{Name: "geo", Info: map[string][]string{
				"region": {"us-west"},
			}},
		},
	}
	ctx := result.WithResult(context.Background(), res)

	rsp, err := (&Server{}).Echo(ctx, &EchoRequest{})
	require.NoError(t, err)
	require.NotNil(t, rsp)

	want := []extTuple{
		{Name: "auth", Key: "subject", Values: []string{"alice"}},
		{Name: "auth", Key: "audience", Values: []string{"a", "b"}},
		{Name: "geo", Key: "region", Values: []string{"us-west"}},
	}
	assert.ElementsMatch(t, want, flattenExts(rsp.Extensions))
}

func TestEcho_EmptyExtensions(t *testing.T) {
	rsp, err := (&Server{}).Echo(context.Background(), &EchoRequest{})
	require.NoError(t, err)
	require.NotNil(t, rsp)
	assert.Nil(t, rsp.Extensions)
}

func flatten(kms []*KeyMultivalue) map[string][]string {
	out := make(map[string][]string, len(kms))
	for _, kv := range kms {
		out[kv.Key] = kv.Values
	}
	return out
}

type extTuple struct {
	Name   string
	Key    string
	Values []string
}

func flattenExts(exts []*ExtendedInfo) []extTuple {
	out := make([]extTuple, 0, len(exts))
	for _, ext := range exts {
		out = append(out, extTuple{
			Name:   ext.Name,
			Key:    ext.Info.Key,
			Values: ext.Info.Values,
		})
	}
	return out
}
