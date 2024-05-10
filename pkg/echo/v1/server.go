package v1

import (
	"context"
	"runtime"
	"runtime/debug"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ripta/netdebug/pkg/echo/result"
)

type Server struct {
	UnimplementedEchoerServer
}

func (s *Server) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	res := result.FromContext(ctx)
	rsp := EchoResponse{
		Query: req.Query,
		Request: &RequestInfo{
			Protocol:   res.Request.Protocol,
			RemoteAddr: res.Request.RemoteAddr,
			Method:     res.Request.Method,
			Uri:        res.Request.URI,
			Header:     buildKeyMultivalue(res.Request.Headers),
			ParsedUrl: &ParsedURL{
				Scheme:   res.Request.ParsedURL.Scheme,
				Host:     res.Request.ParsedURL.Host,
				Path:     res.Request.ParsedURL.Path,
				RawPath:  res.Request.ParsedURL.RawPath,
				RawQuery: res.Request.ParsedURL.RawQuery,
				Query:    buildKeyMultivalue(res.Request.ParsedURL.Query),
			},
		},
		Runtime: &RuntimeInfo{
			GoVersion:     runtime.Version(),
			GoArch:        runtime.GOARCH,
			GoOs:          runtime.GOOS,
			NumCpus:       int64(runtime.NumCPU()),
			NumGoroutines: int64(runtime.NumGoroutine()),
		},
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		rsp.Runtime.MainPath = info.Path
		rsp.Runtime.MainModule = info.Main.Path

		rsp.Runtime.MainVersion = info.Main.Version
		if info.Main.Version == "(devel)" {
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" {
					rsp.Runtime.MainVersion = s.Value
				}
				if s.Key == "vcs.modified" && s.Value == "true" {
					rsp.Runtime.MainVersion += " (dirty)"
				}
			}
		}
	}

	for _, exres := range res.Extensions {
		for _, kv := range buildKeyMultivalue(exres.Info) {
			rsp.Extensions = append(rsp.Extensions, &ExtendedInfo{
				Name: exres.Name,
				Info: kv,
			})
		}
	}

	return &rsp, nil
}

func (s *Server) Status(_ context.Context, req *StatusRequest) (*StatusResponse, error) {
	code := codes.Code(req.ForceGrpcStatus)
	return &StatusResponse{}, status.Error(code, req.Message)
}

func buildKeyMultivalue(md map[string][]string) []*KeyMultivalue {
	kms := []*KeyMultivalue{}
	for k, mv := range md {
		kms = append(kms, &KeyMultivalue{
			Key:    k,
			Values: mv,
		})
	}

	return kms
}
