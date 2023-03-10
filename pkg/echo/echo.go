package echo

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"k8s.io/klog/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"sync"
	"syscall"
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
	klog.V(3).InfoS("serving request", "request_uri", r.RequestURI)
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "Kubernetes details:\n")
	fmt.Fprintf(w, "\tHostname: %s\n", s.Hostname)
	fmt.Fprintf(w, "\tPod name: %s\n", s.PodName)
	fmt.Fprintf(w, "\tNamespace: %s\n", s.PodNamespace)
	fmt.Fprintf(w, "\tNode name: %s\n", s.PodNode)
	fmt.Fprint(w, "\n")

	fmt.Fprint(w, "Request information:\n")
	fmt.Fprintf(w, "\tProtocol: %s\n", r.Proto)
	if r.TLS == nil {
		fmt.Fprint(w, "\tTLS: none\n")
	} else {
		switch r.TLS.Version {
		case tls.VersionTLS10:
			fmt.Fprint(w, "\tTLS: TLSv1.0\n")
		case tls.VersionTLS11:
			fmt.Fprint(w, "\tTLS: TLSv1.1\n")
		case tls.VersionTLS12:
			fmt.Fprint(w, "\tTLS: TLSv1.2\n")
		case tls.VersionTLS13:
			fmt.Fprint(w, "\tTLS: TLSv1.3\n")
		default:
			fmt.Fprintf(w, "\tTLS: unknown (version=%d)\n", r.TLS.Version)
		}
	}
	fmt.Fprintf(w, "\tRemote address: %s\n", r.RemoteAddr)
	fmt.Fprintf(w, "\tMethod: %s\n", r.Method)
	fmt.Fprintf(w, "\tRaw URI: %s\n", r.RequestURI)
	if u := r.URL; u != nil {
		fmt.Fprintf(w, "\t\tPath: %s\n", u.Path)
		fmt.Fprintf(w, "\t\tQuery: %s\n", u.RawQuery)
	}
	fmt.Fprint(w, "\n")

	names := []string{}
	for hk := range r.Header {
		names = append(names, hk)
	}
	sort.Strings(names)

	fmt.Fprint(w, "Request headers:\n")
	for _, hk := range names {
		for _, hv := range r.Header[hk] {
			fmt.Fprintf(w, "\t%s: %s\n", hk, hv)
		}
	}
	fmt.Fprint(w, "\n")

	fmt.Fprint(w, "Runtime information:\n")
	fmt.Fprintf(w, "\tVersion: %s\n", runtime.Version())
	fmt.Fprintf(w, "\tArch/OS: %s/%s\n", runtime.GOARCH, runtime.GOOS)
	fmt.Fprintf(w, "\tNumber of CPUs: %d\n", runtime.NumCPU())
	fmt.Fprintf(w, "\tNumber of goroutines: %d\n", runtime.NumGoroutine())
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Fprintf(w, "\tApp main module: %s\n", info.Main.Path)
		fmt.Fprintf(w, "\tApp main version: %s\n", info.Main.Version)
	}
	fmt.Fprint(w, "\n")
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
