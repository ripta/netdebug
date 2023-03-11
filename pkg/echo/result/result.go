package result

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/url"
	"sort"
)

type Result struct {
	Kubernetes KubernetesResult `json:"kubernetes"`
	Request    RequestResult    `json:"request"`
	Runtime    RuntimeResult    `json:"runtime"`
}

var _ io.WriterTo = Result{}

func (r Result) WriteTo(w io.Writer) (int64, error) {
	buf := &bytes.Buffer{}

	fmt.Fprint(buf, "Kubernetes details:\n")
	fmt.Fprintf(buf, "\tHostname: %s\n", r.Kubernetes.Hostname)
	fmt.Fprintf(buf, "\tPod name: %s\n", r.Kubernetes.PodName)
	fmt.Fprintf(buf, "\tNamespace: %s\n", r.Kubernetes.PodNamespace)
	fmt.Fprintf(buf, "\tNode name: %s\n", r.Kubernetes.PodNode)
	fmt.Fprint(buf, "\n")

	fmt.Fprint(buf, "Request information:\n")
	fmt.Fprintf(buf, "\tProtocol: %s\n", r.Request.Protocol)

	fmt.Fprintf(buf, "\tRemote address: %s\n", r.Request.RemoteAddr)
	fmt.Fprintf(buf, "\tMethod: %s\n", r.Request.Method)
	fmt.Fprintf(buf, "\tRaw URI: %s\n", r.Request.URI)
	fmt.Fprintf(buf, "\tRaw Query: %s\n", r.Request.ParsedURL.RawQuery)
	fmt.Fprint(buf, "\n")

	names := []string{}
	for hk := range r.Request.Headers {
		names = append(names, hk)
	}
	sort.Strings(names)

	fmt.Fprint(buf, "Request headers:\n")
	for _, hk := range names {
		for _, hv := range r.Request.Headers[hk] {
			fmt.Fprintf(buf, "\t%s: %s\n", hk, hv)
		}
	}
	fmt.Fprint(buf, "\n")

	fmt.Fprint(buf, "Runtime information:\n")
	fmt.Fprintf(buf, "\tVersion: %s\n", r.Runtime.GoVersion)
	fmt.Fprintf(buf, "\tArch/OS: %s/%s\n", r.Runtime.GoArch, r.Runtime.GoOS)
	fmt.Fprintf(buf, "\tNumber of CPUs: %d\n", r.Runtime.NumCPUs)
	fmt.Fprintf(buf, "\tNumber of goroutines: %d\n", r.Runtime.NumGoroutines)
	fmt.Fprintf(buf, "\tApp main module: %s\n", r.Runtime.MainModule)
	fmt.Fprintf(buf, "\tApp main path: %s\n", r.Runtime.MainPath)
	fmt.Fprintf(buf, "\tApp main version: %s\n", r.Runtime.MainVersion)
	fmt.Fprint(buf, "\n")

	return buf.WriteTo(w)
}

type KubernetesResult struct {
	Hostname     string `json:"hostname"`
	PodName      string `json:"pod_name"`
	PodNamespace string `json:"pod_namespace"`
	PodNode      string `json:"pod_node"`
}

type RequestResult struct {
	Protocol   string              `json:"protocol"`
	TLSVersion string              `json:"tls_version"`
	RemoteAddr string              `json:"remote_addr"`
	Method     string              `json:"method"`
	URI        string              `json:"uri"`
	ParsedURL  ParsedURL           `json:"parsed_url"`
	Headers    map[string][]string `json:"headers"`
}

type ParsedURL struct {
	Scheme   string     `json:"scheme"`
	Host     string     `json:"host"`
	Path     string     `json:"path"`
	RawPath  string     `json:"raw_path"`
	RawQuery string     `json:"raw_query"`
	Query    url.Values `json:"query"`
}

type RuntimeResult struct {
	GoVersion     string `json:"go_version"`
	GoArch        string `json:"go_arch"`
	GoOS          string `json:"go_os"`
	NumCPUs       int    `json:"num_cpus"`
	NumGoroutines int    `json:"num_goroutines"`
	MainModule    string `json:"main_module"`
	MainPath      string `json:"main_path"`
	MainVersion   string `json:"main_version"`
}

func TLSVersion(cs *tls.ConnectionState) string {
	if cs == nil {
		return "none"
	}
	switch cs.Version {
	case tls.VersionTLS10:
		return "TLSv1.0"
	case tls.VersionTLS11:
		return "TLSv1.1"
	case tls.VersionTLS12:
		return "TLSv1.2"
	case tls.VersionTLS13:
		return "TLSv1.3"
	default:
	}
	return fmt.Sprintf("unknown (version=%d)", cs.Version)
}
