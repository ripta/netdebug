package echo

import (
	"fmt"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
)

type Server struct {
	Hostname     string
	ListenHost   string
	ListenPort   int
	PodName      string
	PodNamespace string
	PodNode      string
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

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", http.NotFound)
	mux.HandleFunc("/healthz", s.healthzHandler)
	mux.HandleFunc("/", s.echoHandler)

	klog.InfoS("listening for HTTP requests", "host", s.ListenHost, "port", s.ListenPort)
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.ListenHost, s.ListenPort), mux)
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
