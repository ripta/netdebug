package app

import (
	"context"
	"errors"
	goflag "flag"
	"strings"

	"github.com/ripta/rt/pkg/version"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

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

	cmd.AddCommand(newDNSCommand())
	cmd.AddCommand(newEchoCommand())
	cmd.AddCommand(newListenCommand())
	cmd.AddCommand(newSendCommand())
	cmd.AddCommand(version.NewCommand())

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

	cmd.Flags().Float64Var(&s.OtelSampleRate, "otel-sample-rate", s.OtelSampleRate, "OpenTelemetry sample rate, between 0 (never) and 1 (always) inclusive")
	cmd.Flags().StringVar(&s.OtelGCPProjectID, "otel-gcp-project-id", s.OtelGCPProjectID, "OpenTelemetry Exporter's GCP Project ID")

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
