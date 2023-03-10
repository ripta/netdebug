package result

import (
	"net/http"
	"runtime"
	"runtime/debug"
)

func GetRequestResult(r *http.Request) RequestResult {
	rr := RequestResult{
		Protocol:   r.Proto,
		TLSVersion: TLSVersion(r.TLS),
		RemoteAddr: r.RemoteAddr,
		Method:     r.Method,
		URI:        r.RequestURI,
		Headers:    r.Header,
	}

	if u := r.URL; u != nil {
		rr.ParsedURL = ParsedURL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     u.Path,
			RawPath:  u.RawPath,
			RawQuery: u.RawQuery,
			Query:    u.Query(),
		}
	} else {
		rr.ParsedURL.Path = r.RequestURI
	}

	return rr
}

func GetRuntimeResult() RuntimeResult {
	rt := RuntimeResult{
		GoVersion:     runtime.Version(),
		GoArch:        runtime.GOARCH,
		GoOS:          runtime.GOOS,
		NumCPUs:       runtime.NumCPU(),
		NumGoroutines: runtime.NumGoroutine(),
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		rt.MainPath = info.Path
		rt.MainModule = info.Main.Path

		rt.MainVersion = info.Main.Version
		if info.Main.Version == "(devel)" {
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" {
					rt.MainVersion = s.Value
				}
				if s.Key == "vcs.modified" && s.Value == "true" {
					rt.MainVersion += " (dirty)"
				}
			}
		}
	}

	return rt
}
