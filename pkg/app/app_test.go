package app

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findSubcommand(root *cobra.Command, name string) *cobra.Command {
	for _, sub := range root.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}

func TestNewRootCommand_WiresAllSubcommands(t *testing.T) {
	root := NewRootCommand()
	for _, name := range []string{"dns", "echo", "listen", "send"} {
		t.Run(name, func(t *testing.T) {
			assert.NotNilf(t, findSubcommand(root, name), "subcommand %q not registered", name)
		})
	}
}

type subcommandFlagTest struct {
	Name    string
	Command string
	Flag    string
}

var subcommandFlagTests = []subcommandFlagTest{
	{Name: "dns_type", Command: "dns", Flag: "type"},
	{Name: "dns_server_addr", Command: "dns", Flag: "dns-server-addr"},
	{Name: "echo_host", Command: "echo", Flag: "host"},
	{Name: "echo_port", Command: "echo", Flag: "port"},
	{Name: "echo_mode", Command: "echo", Flag: "mode"},
	{Name: "echo_tls_autogenerate", Command: "echo", Flag: "tls-autogenerate"},
	{Name: "echo_tls_key_file", Command: "echo", Flag: "tls-key-file"},
	{Name: "echo_tls_cert_file", Command: "echo", Flag: "tls-cert-file"},
	{Name: "echo_jwt_header_name", Command: "echo", Flag: "jwt-header-name"},
	{Name: "echo_jwt_jwks_url", Command: "echo", Flag: "jwt-jwks-url"},
	{Name: "echo_jwt_issuer_url", Command: "echo", Flag: "jwt-issuer-url"},
	{Name: "echo_jwt_audience", Command: "echo", Flag: "jwt-audience"},
	{Name: "echo_jwt_signing_algorithms", Command: "echo", Flag: "jwt-signing-algorithms"},
	{Name: "listen_host", Command: "listen", Flag: "host"},
	{Name: "listen_network", Command: "listen", Flag: "network"},
	{Name: "listen_port", Command: "listen", Flag: "port"},
	{Name: "send_network", Command: "send", Flag: "network"},
	{Name: "send_address", Command: "send", Flag: "address"},
}

func TestNewRootCommand_SubcommandFlags(t *testing.T) {
	root := NewRootCommand()
	for _, tc := range subcommandFlagTests {
		t.Run(tc.Name, func(t *testing.T) {
			sub := findSubcommand(root, tc.Command)
			require.NotNilf(t, sub, "subcommand %q not registered", tc.Command)
			assert.NotNilf(t, sub.Flags().Lookup(tc.Flag), "flag --%s missing on subcommand %q", tc.Flag, tc.Command)
		})
	}
}

func TestNew_ReturnsRootAndCleanup(t *testing.T) {
	cmd, cleanup := New()
	require.NotNil(t, cmd)
	require.NotNil(t, cleanup)

	// klog.InitFlags installs "v" (verbosity) among others; presence on the
	// persistent flag set confirms AddGoFlagSet wired the klog flags.
	assert.NotNil(t, cmd.PersistentFlags().Lookup("v"), "klog --v flag missing on root persistent flags")
}
