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
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"k8s.io/klog/v2"

	"github.com/ripta/netdebug/pkg/echo/result"
	"github.com/ripta/netdebug/pkg/echo/v1"
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

	ns, err := os.ReadFile("/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		ns = []byte(fmt.Sprintf("(error: %v)", err))
	}

	return &Server{
		Hostname:     hostname,
		ListenPort:   8080,
		Mode:         ServerModeHTTP,
		PodName:      getEnvOrDefault("POD_NAME", "($POD_NAME unset)"),
		PodNamespace: string(ns),
		PodNode:      getEnvOrDefault("NODE_NAME", "($NODE_NAME unset)"),
	}
}

func (s *Server) Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if s.TLSAutogen && s.TLSCertPath != "" && s.TLSKeyPath != "" {
		return errors.New("--tls-autogenerate cannot be combined with --tls-key-file and --tls-cert-file")
	} else if (s.TLSCertPath != "") != (s.TLSKeyPath != "") {
		return errors.New("--tls-key-file and --tls-cert-file must both be empty or both be specified")
	} else if s.Mode != ServerModeHTTP && s.TLSKeyPath == "" && !s.TLSAutogen {
		return errors.New("--mode=grpc currently requires TLS to be enabled (--tls-{cert,key}-file or --tls-autogenerate)")
	}

	if s.TLSAutogen {
		tlsDir, err := os.MkdirTemp("", "netdebug-echo-tls.*")
		if err != nil {
			return err
		}

		klog.InfoS("generating self-signed certificates", "tls_dir", tlsDir)
		defer func() {
			klog.InfoS("cleaning up self-signed certificates", "tls_dir", tlsDir)
			os.RemoveAll(tlsDir)
		}()

		klog.Info("generating CA certificate")
		caCert, err := generateCACert()
		if err != nil {
			return fmt.Errorf("generate CA cert: %w", err)
		}

		klog.Info("generating server certificate")
		servCert, err := generateServerCert(caCert)
		if err != nil {
			return fmt.Errorf("generating server cert: %w", err)
		}

		caCertFile := filepath.Join(tlsDir, "ca.crt")
		if err := os.WriteFile(caCertFile, caCert.CertPEM, 0o644); err != nil {
			return fmt.Errorf("writing CA cert: %w", err)
		}

		caKeyFile := filepath.Join(tlsDir, "ca.key")
		if err := os.WriteFile(caKeyFile, caCert.PrivatePEM, 0o600); err != nil {
			return fmt.Errorf("writing CA key: %w", err)
		}

		servCertFile := filepath.Join(tlsDir, "server.crt")
		if err := os.WriteFile(servCertFile, servCert.CertPEM, 0o644); err != nil {
			return fmt.Errorf("writing server cert: %w", err)
		}

		servKeyFile := filepath.Join(tlsDir, "server.key")
		if err := os.WriteFile(servKeyFile, servCert.PrivatePEM, 0o600); err != nil {
			return fmt.Errorf("writing server key: %w", err)
		}

		s.TLSCertPath = servCertFile
		s.TLSKeyPath = servKeyFile
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", http.NotFound)
	mux.HandleFunc("/healthz", s.healthzHandler)
	if s.Mode == ServerModeGRPC || s.Mode == ServerModeBoth {
		klog.InfoS("initializing gRPC handler")

		gs := grpc.NewServer()
		v1.RegisterEchoerServer(gs, &v1.Server{})
		reflection.Register(gs)

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if s.Mode == ServerModeBoth && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
				s.echoHandler(w, r)
				return
			}

			res := s.getResultFromRequest(r)
			klog.V(3).InfoS("serving gRPC request", "request_uri", r.RequestURI, "remote_addr", r.RemoteAddr)

			ctx := r.Context()
			gs.ServeHTTP(w, r.WithContext(result.WithResult(ctx, res)))
		})
	} else {
		klog.InfoS("initializing HTTP handler")
		mux.HandleFunc("/", s.echoHandler)
	}

	addr := fmt.Sprintf("%s:%d", s.ListenHost, s.ListenPort)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

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
			klog.InfoS("listening for HTTP requests with TLS", "addr", addr)
			if err := server.ListenAndServeTLS(s.TLSCertPath, s.TLSKeyPath); err != nil && err != http.ErrServerClosed {
				klog.ErrorS(err, "after serving with TLS")
			}
			return
		}

		klog.InfoS("listening for HTTP requests", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.ErrorS(err, "after serving")
		}
	}()

	wg.Wait()

	klog.Info("shut down complete")
	return nil
}

func (s *Server) getResultFromRequest(r *http.Request) result.Result {
	return result.Result{
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
	w.Write([]byte("OK\n"))
}

func getEnvOrDefault(envName, defaultValue string) string {
	env, ok := os.LookupEnv(envName)
	if ok {
		return env
	}
	return defaultValue
}
