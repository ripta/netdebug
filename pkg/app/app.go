package app

import (
	"context"
	"errors"
	goflag "flag"
	"strings"

	"github.com/ripta/rt/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"github.com/ripta/netdebug/pkg/bench"
	"github.com/ripta/netdebug/pkg/dns"
	"github.com/ripta/netdebug/pkg/echo"
	"github.com/ripta/netdebug/pkg/echo/extensions"
	"github.com/ripta/netdebug/pkg/listen"
	"github.com/ripta/netdebug/pkg/send"
)

type CleanupFunc func()

func New() (*cobra.Command, CleanupFunc) {
	cmd := NewRootCommand()

	f := goflag.NewFlagSet("klog", goflag.ContinueOnError)
	klog.InitFlags(f)
	cmd.PersistentFlags().AddGoFlagSet(f)

	return cmd, klog.Flush
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "netdebug",
		Short:         "A collection of network debugging tools",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	_ = cmd.MarkFlagFilename("config", "yaml", "json")

	cmd.AddCommand(newBenchCommand())
	cmd.AddCommand(newDNSCommand())
	cmd.AddCommand(newEchoCommand())
	cmd.AddCommand(newListenCommand())
	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(version.NewCommand())

	return cmd
}

func newBenchCommand() *cobra.Command {
	b := bench.New()
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "Benchmark a gRPC echo server",
		Long: `Drive load against a gRPC echo server with configurable payload shape,
compression, and connection model.

Conn-model semantics differ depending on whether a service mesh is in the
path. Without a mesh, kube-proxy load-balances per connection (L4), so
"shared" pins every worker to one backend while "per-worker" gives N
connections through the LB. With a mesh, the sidecar load-balances per
request (L7), so "shared" still spreads work across backends. "per-request"
forces a fresh TCP+TLS+HTTP/2 handshake on every call regardless of mesh,
which is the way to surface connection-establishment cost.

Interpreting the output. The summary breaks total RPC time into "server"
(from the EchoResponse.server_duration_ns field set by the handler) and
"network" (total - server, which is wire time plus any proxy or mesh hops).
"upstream" appears only when the response carries the
x-envoy-upstream-service-time header that Envoy sets when an Istio sidecar
fronts the backend; it covers the server-side proxy plus handler, so
"network - upstream" isolates the client-proxy plus on-wire portion.

Backends are grouped by kubernetes.pod_name from the response, falling
back to hostname and then the resolved peer address; the per-row "source"
identifies which step in the chain produced the key. Backend skew is the
ratio of max to min across backends, reported separately for request
count and p99 latency; a large ratio is a single-replica problem, such as
an overloaded pod, slow disk, or stuck retry loop. "Errors by code"
buckets failures by gRPC status code and surfaces the top distinct
messages per bucket.

Service mesh notes. Under Envoy/Istio, the upstream block is present and
gives a direct read on server-side proxy plus handler time.
Linkerd2-proxy does not emit an equivalent response-side timing header,
so the upstream block reads "n/a" under linkerd even when a sidecar is in
the path; the decomposition stops at total - server = client_proxy +
network + server_proxy with no in-band split between those three.
linkerd2-proxy's latency signal lives on the proxy's :4191/metrics
Prometheus endpoint, which bench does not scrape.`,
		Example: "netdebug bench --target=127.0.0.1:8080 --payload=embedding-float --embedding-dim=1024 --concurrency=4 --duration=10s",
		RunE:    runAdapter(b.Run),
	}

	cmd.Flags().StringVarP(&b.Target, "target", "t", b.Target, "Target address (host:port) of the echo server")
	cmd.Flags().BoolVar(&b.Plaintext, "plaintext", b.Plaintext, "Use plaintext gRPC instead of TLS")
	cmd.Flags().BoolVar(&b.TLSInsecureSkipVerify, "tls-insecure-skip-verify", b.TLSInsecureSkipVerify, "Skip TLS certificate verification; only meaningful without --plaintext")
	cmd.Flags().IntVarP(&b.Concurrency, "concurrency", "c", b.Concurrency, "Number of concurrent workers")
	cmd.Flags().DurationVarP(&b.Duration, "duration", "d", b.Duration, "Duration of the benchmark run")
	cmd.Flags().VarP(&b.Payload, "payload", "p", `Payload mix, e.g. "embedding-float" or "embedding-float:50,embedding-bytes:50"`)
	cmd.Flags().IntVar(&b.EmbeddingDim, "embedding-dim", b.EmbeddingDim, "Dimensions for embedding-float and embedding-bytes payload shapes")
	cmd.Flags().IntVar(&b.BytesSize, "bytes-size", b.BytesSize, "Size in bytes for the bytes payload shape")
	cmd.Flags().IntVar(&b.StringLen, "string-len", b.StringLen, "Length in characters for the string payload shape")
	cmd.Flags().StringVar(&b.Compression, "compression", b.Compression, "Compression codec: identity, gzip, snappy, zstd")
	cmd.Flags().StringVar(&b.ConnModel, "conn-model", b.ConnModel, "Connection model: per-worker, shared, per-request")
	cmd.Flags().StringVar(&b.OutputFormat, "output", b.OutputFormat, "Output format: human, json")
	cmd.Flags().StringToStringVar(&b.Headers, "header", b.Headers, "Repeatable header (key=value) attached to every outgoing gRPC request")
	cmd.Flags().StringToStringVar(&b.Labels, "label", b.Labels, "Repeatable label (key=value) included verbatim in the JSON summary and the human header")

	return cmd
}

func newDNSCommand() *cobra.Command {
	d := dns.New()
	cmd := &cobra.Command{
		Use:     "dns",
		Short:   "Perform DNS query",
		Example: "netdebug dns -t mx r8y.org",
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) < 1 || args[0] == "" {
				return errors.New("name to query must not be empty")
			}

			d.QueryName = args[0]
			if !strings.HasPrefix(d.QueryName, ".") {
				d.QueryName += "."
			}

			return d.Run()
		},
	}

	cmd.Flags().StringVarP(&d.ServerAddress, "dns-server-addr", "d", d.ServerAddress, "DNS server address with port, e.g., 127.0.0.1:53, [::1]:53")
	cmd.Flags().VarP(&d.QueryType, "type", "t", "Query type, e.g., mx, cname")

	return cmd
}

func newEchoCommand() *cobra.Command {
	s := echo.New()
	jwtConf := extensions.JWTConfig{}

	cmd := &cobra.Command{
		Use:     "echo",
		Short:   "HTTP echo server",
		Example: "netdebug echo -v=3",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if jwtConf.HeaderName != "" {
				ext, err := extensions.JWT(jwtConf)
				if err != nil {
					return err
				}

				s.InstallExtension(ext)
			}

			return s.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&s.ListenHost, "host", "H", s.ListenHost, "Host to listen on")
	cmd.Flags().IntVarP(&s.ListenPort, "port", "p", s.ListenPort, "Server port to listen on")
	cmd.Flags().BoolVar(&s.TLSAutogen, "tls-autogenerate", s.TLSAutogen, "Automatically generate TLS key and cert")
	cmd.Flags().StringVar(&s.TLSKeyPath, "tls-key-file", s.TLSKeyPath, "Path to TLS key")
	cmd.Flags().StringVar(&s.TLSCertPath, "tls-cert-file", s.TLSCertPath, "Path to TLS cert")

	cmd.Flags().StringVar(&jwtConf.HeaderName, "jwt-header-name", jwtConf.HeaderName, "JWT header name")
	cmd.Flags().StringVar(&jwtConf.JWKSURL, "jwt-jwks-url", jwtConf.JWKSURL, "JWT JWKS URL")
	cmd.Flags().StringVar(&jwtConf.IssuerURL, "jwt-issuer-url", jwtConf.IssuerURL, "JWT issuer URL")
	cmd.Flags().StringVar(&jwtConf.Audience, "jwt-audience", jwtConf.Audience, "JWT audience")
	cmd.Flags().StringSliceVar(&jwtConf.SigningAlgorithms, "jwt-signing-algorithms", jwtConf.SigningAlgorithms, "JWT supported signing algorithms")

	echo.ServerModeVar(cmd.Flags(), &s.Mode, "mode", "Server mode: http, grpc, grpc+http")

	return cmd
}

func newListenCommand() *cobra.Command {
	l := listen.New()
	cmd := &cobra.Command{
		Use:     "listen",
		Short:   "Listen for connection",
		Example: "netdebug listen -p 15921",
		RunE:    runAdapter(l.Run),
	}

	cmd.Flags().StringVarP(&l.Host, "host", "H", l.Host, "Host to listen on")
	cmd.Flags().StringVarP(&l.Network, "network", "n", l.Network, "Network to listen on, one of: tcp, tcp4, tcp6, unix, or unixpacket")
	cmd.Flags().IntVarP(&l.Port, "port", "p", l.Port, "Port number to listen on (0 = first available)")

	return cmd
}

func newSendCommand() *cobra.Command {
	s := send.New()
	cmd := &cobra.Command{
		Use:     "send",
		Short:   "Send packet",
		Example: "date | netdebug send -a 192.168.11.1:15921",
		RunE:    runAdapter(s.Run),
	}

	cmd.Flags().StringVarP(&s.Network, "network", "n", s.Network, "Network to send on, one of: tcp, tcp4, tcp6, unix, or unixpacket")
	cmd.Flags().StringVarP(&s.Address, "address", "a", s.Address, "Address to send to")

	return cmd
}

func runAdapter(f func(ctx context.Context) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return f(cmd.Context())
	}
}
