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
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"

	"k8s.io/klog/v2"
)

type Server struct {
	Hostname     string
	ListenHost   string
	ListenPort   int
	PodName      string
	PodNamespace string
	PodNode      string
	TLSAutogen   bool
	TLSCertPath  string
	TLSKeyPath   string
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
	mux.HandleFunc("/", s.echoHandler)

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

func (s *Server) echoHandler(w http.ResponseWriter, r *http.Request) {
	res := Result{
		Kubernetes: KubernetesResult{
			Hostname:     s.Hostname,
			PodName:      s.PodName,
			PodNamespace: s.PodNamespace,
			PodNode:      s.PodNode,
		},
		Request: RequestResult{
			Protocol:   r.Proto,
			TLSVersion: tlsVersion(r.TLS),
			RemoteAddr: r.RemoteAddr,
			Method:     r.Method,
			URI:        r.RequestURI,
			Headers:    r.Header,
		},
		Runtime: RuntimeResult{
			GoVersion:     runtime.Version(),
			GoArch:        runtime.GOARCH,
			GoOS:          runtime.GOOS,
			NumCPUs:       runtime.NumCPU(),
			NumGoroutines: runtime.NumGoroutine(),
		},
	}

	if u := r.URL; u != nil {
		res.Request.ParsedURL = ParsedURL{
			Scheme:   u.Scheme,
			Host:     u.Host,
			Path:     u.Path,
			RawPath:  u.RawPath,
			RawQuery: u.RawQuery,
			Query:    u.Query(),
		}
	} else {
		res.Request.ParsedURL.Path = r.RequestURI
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		res.Runtime.MainPath = info.Path
		res.Runtime.MainModule = info.Main.Path

		res.Runtime.MainVersion = info.Main.Version
		if info.Main.Version == "(devel)" {
			for _, s := range info.Settings {
				if s.Key == "vcs.revision" {
					res.Runtime.MainVersion = s.Value
				}
				if s.Key == "vcs.modified" && s.Value == "true" {
					res.Runtime.MainVersion += " (dirty)"
				}
			}
		}
	}

	klog.V(3).InfoS("serving request", "request_uri", r.RequestURI, "remote_addr", r.RemoteAddr)
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
