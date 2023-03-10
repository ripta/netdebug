package v1

import (
	"context"
	"runtime"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Server struct {
	UnimplementedEchoerServer
}

func (s *Server) Echo(ctx context.Context, req *EchoRequest) (*EchoResponse, error) {
	sts := grpc.ServerTransportStreamFromContext(ctx)

	addr := ""
	if p, ok := peer.FromContext(ctx); ok {
		addr = p.Addr.String()
	}

	rsp := EchoResponse{
		Query: req.Query,
		Request: &RequestInfo{
			Protocol:   "",
			RemoteAddr: addr,
			Method:     "",
			Uri:        sts.Method(),
			ParsedUrl:  nil,
			Header:     buildKeyMultivalueFromContext(ctx),
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

	return &rsp, nil
}

func buildKeyMultivalueFromContext(ctx context.Context) []*KeyMultivalue {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	kms := []*KeyMultivalue{}
	for k, mv := range md {
		kms = append(kms, &KeyMultivalue{
			Key:    k,
			Values: mv,
		})
	}

	return kms
}
