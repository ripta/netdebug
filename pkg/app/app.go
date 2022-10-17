package app

import (
	goflag "flag"
	"github.com/ripta/netdebug/pkg/echo"

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

	return cmd
}

func newEchoCommand() *cobra.Command {
	s := echo.New()
	cmd := &cobra.Command{
		Use:   "echo",
		Short: "HTTP echo server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return s.Run()
		},
	}

	cmd.Flags().IntVarP(&s.Port, "port", "p", s.Port, "Server port to listen on")

	return cmd
}
