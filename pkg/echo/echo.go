package echo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/pflag"
	"github.com/thediveo/enumflag/v2"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog/v2"

	"github.com/ripta/netdebug/pkg/echo/result"
	v1 "github.com/ripta/netdebug/pkg/echo/v1"
)

type Server struct {
	Hostname     string
	ListenHost   string
	ListenPort   int
	Mode         ServerMode
	PodName      string
	PodNamespace string
	PodNode      string
	TLSAutogen   bool
	TLSCertPath  string
	TLSKeyPath   string
	Extensions   result.Extensions
}

type ServerMode enumflag.Flag

const (
	ServerModeHTTP ServerMode = iota
	ServerModeGRPC
	ServerModeBoth
)

var ServerModeOptions = map[ServerMode][]string{
	ServerModeHTTP: {"", "http"},
	ServerModeGRPC: {"grpc"},
	ServerModeBoth: {"grpc+http", "both"},
}

func ServerModeVar(flags *pflag.FlagSet, sm *ServerMode, name, usage string) {
	f := enumflag.New(sm, name, ServerModeOptions, enumflag.EnumCaseInsensitive)
	flags.Var(f, name, usage)
}

func New() *Server {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("(error: %v)", err)
	}

	ns, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		ns = []byte(fmt.Sprintf("(error: %v)", err))
	}

	return &Server{
		Hostname:     hostname,
		ListenPort:   8080,
		Mode:         ServerModeBoth,
		PodName:      getEnvOrDefault("POD_NAME", "($POD_NAME unset)"),
		PodNamespace: string(ns),
		PodNode:      getEnvOrDefault("NODE_NAME", "($NODE_NAME unset)"),
		Extensions:   []result.ExtensionFunc{},
	}
}

func (s *Server) InstallExtension(fn result.ExtensionFunc) {
	s.Extensions = append(s.Extensions, fn)
}

func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := s.validateTLSFlags(); err != nil {
		return err
	}

	if s.TLSAutogen {
		cleanup, err := s.setupTLSAutogen()
		if err != nil {
			return err
		}
		defer cleanup()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", http.NotFound)
	mux.HandleFunc("/healthz", s.healthzHandler)

	switch s.Mode {
	case ServerModeHTTP:
		s.installHTTPHandler(mux)
	case ServerModeGRPC:
		s.installGRPCHandler(mux)
	case ServerModeBoth:
		s.installBothHandler(mux)
	}

	addr := fmt.Sprintf("%s:%d", s.ListenHost, s.ListenPort)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	s.serve(ctx, server)
	return nil
}

func (s *Server) validateTLSFlags() error {
	if s.TLSAutogen && s.TLSCertPath != "" && s.TLSKeyPath != "" {
		return errors.New("--tls-autogenerate cannot be combined with --tls-key-file and --tls-cert-file")
	}
	if (s.TLSCertPath != "") != (s.TLSKeyPath != "") {
		return errors.New("--tls-key-file and --tls-cert-file must both be empty or both be specified")
	}
	return nil
}

func (s *Server) setupTLSAutogen() (func(), error) {
	tlsDir, err := os.MkdirTemp("", "netdebug-echo-tls.*")
	if err != nil {
		return nil, err
	}

	klog.InfoS("generating self-signed certificates", "tls_dir", tlsDir)
	cleanup := func() {
		klog.InfoS("cleaning up self-signed certificates", "tls_dir", tlsDir)
		if err := os.RemoveAll(tlsDir); err != nil {
			klog.ErrorS(err, "removing TLS directory", "tls_dir", tlsDir)
		}
	}

	klog.Info("generating CA certificate")
	caCert, err := generateCACert()
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("generate CA cert: %w", err)
	}

	klog.Info("generating server certificate")
	servCert, err := generateServerCert(caCert)
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("generating server cert: %w", err)
	}

	caCertFile := filepath.Join(tlsDir, "ca.crt")
	if err := os.WriteFile(caCertFile, caCert.CertPEM, 0o644); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing CA cert: %w", err)
	}

	caKeyFile := filepath.Join(tlsDir, "ca.key")
	if err := os.WriteFile(caKeyFile, caCert.PrivatePEM, 0o600); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing CA key: %w", err)
	}

	servCertFile := filepath.Join(tlsDir, "server.crt")
	if err := os.WriteFile(servCertFile, servCert.CertPEM, 0o644); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing server cert: %w", err)
	}

	servKeyFile := filepath.Join(tlsDir, "server.key")
	if err := os.WriteFile(servKeyFile, servCert.PrivatePEM, 0o600); err != nil {
		cleanup()
		return nil, fmt.Errorf("writing server key: %w", err)
	}

	s.TLSCertPath = servCertFile
	s.TLSKeyPath = servKeyFile
	return cleanup, nil
}

func (s *Server) newGRPCServer() *grpc.Server {
	gs := grpc.NewServer()
	v1.RegisterEchoerServer(gs, &v1.Server{})
	reflection.Register(gs)

	gh := health.NewServer()
	gh.SetServingStatus(v1.Echoer_Echo_FullMethodName, grpc_health_v1.HealthCheckResponse_SERVING)
	gh.SetServingStatus(v1.Echoer_Status_FullMethodName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(gs, gh)

	return gs
}

func (s *Server) installHTTPHandler(mux *http.ServeMux) {
	klog.InfoS("initializing HTTP handler")
	mux.HandleFunc("/", s.echoHandler)
}

func (s *Server) installGRPCHandler(mux *http.ServeMux) {
	klog.InfoS("initializing gRPC handler")
	gs := s.newGRPCServer()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res := s.getResultFromRequest(r)
		klog.V(3).InfoS("serving gRPC request", "request_uri", r.RequestURI, "remote_addr", r.RemoteAddr)

		ctx := r.Context()
		gs.ServeHTTP(w, r.WithContext(result.WithResult(ctx, res)))
	})
}

func (s *Server) installBothHandler(mux *http.ServeMux) {
	klog.InfoS("initializing gRPC handler")
	gs := s.newGRPCServer()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			s.echoHandler(w, r)
			return
		}

		res := s.getResultFromRequest(r)
		klog.V(3).InfoS("serving gRPC request", "request_uri", r.RequestURI, "remote_addr", r.RemoteAddr)

		ctx := r.Context()
		gs.ServeHTTP(w, r.WithContext(result.WithResult(ctx, res)))
	})
}

func (s *Server) serve(ctx context.Context, server *http.Server) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		klog.InfoS("waiting for shut down signal")
		<-ctx.Done()

		klog.InfoS("shutting down")
		if err := server.Shutdown(context.Background()); err != nil {
			klog.ErrorS(err, "during shutdown")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if s.TLSKeyPath != "" && s.TLSCertPath != "" {
			klog.InfoS("listening for HTTP requests with TLS", "addr", server.Addr)
			if err := server.ListenAndServeTLS(s.TLSCertPath, s.TLSKeyPath); err != nil && err != http.ErrServerClosed {
				klog.ErrorS(err, "after serving with TLS")
			}
			return
		}

		klog.InfoS("listening for HTTP requests", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.ErrorS(err, "after serving")
		}
	}()

	wg.Wait()

	klog.Info("shut down complete")
}

func (s *Server) getResultFromRequest(r *http.Request) result.Result {
	exts := []result.ExtensionResult{}
	for _, fn := range s.Extensions {
		e, err := fn(r.Clone(context.Background()))
		if err != nil {
			klog.ErrorS(err, "extension error")
		}
		exts = append(exts, e...)
	}

	return result.Result{
		Extensions: s.Extensions.GetResult(r),
		Kubernetes: result.KubernetesResult{
			Hostname:     s.Hostname,
			PodName:      s.PodName,
			PodNamespace: s.PodNamespace,
			PodNode:      s.PodNode,
		},
		Request: result.GetRequestResult(r),
		Runtime: result.GetRuntimeResult(),
	}
}

func (s *Server) echoHandler(w http.ResponseWriter, r *http.Request) {
	res := s.getResultFromRequest(r)

	klog.V(3).InfoS("serving HTTP request", "request_uri", r.RequestURI, "remote_addr", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)

	if strings.HasSuffix(res.Request.ParsedURL.Path, ".json") || strings.Contains(r.Header.Get("Accept"), "application/json") {
		if err := json.NewEncoder(w).Encode(res); err != nil {
			klog.ErrorS(err, "encoding JSON output")
		}

		return
	}

	if _, err := res.WriteTo(w); err != nil {
		klog.ErrorS(err, "writing text output")
	}
}

func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK\n")); err != nil {
		klog.ErrorS(err, "writing healthz response")
	}
}

func getEnvOrDefault(envName, defaultValue string) string {
	env, ok := os.LookupEnv(envName)
	if ok {
		return env
	}
	return defaultValue
}
