package app

import (
	goflag "flag"
	"github.com/ripta/netdebug/pkg/echo"
	"github.com/ripta/netdebug/pkg/listen"
	"github.com/ripta/netdebug/pkg/send"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
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
		Use:   "netdebug",
		Short: "A collection of network debugging tools",
	}

	_ = cmd.MarkFlagFilename("config", "yaml", "json")

	cmd.AddCommand(newEchoCommand())
	cmd.AddCommand(newListenCommand())
	cmd.AddCommand(newSendCommand())

	return cmd
}

func newEchoCommand() *cobra.Command {
	s := echo.New()
	cmd := &cobra.Command{
		Use:     "echo",
		Short:   "HTTP echo server",
		Example: "netdebug echo -v=3",
		RunE:    runAdapter(s.Run),
	}

	cmd.Flags().IntVarP(&s.Port, "port", "p", s.Port, "Server port to listen on")

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

func runAdapter(f func() error) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return f()
	}
}
