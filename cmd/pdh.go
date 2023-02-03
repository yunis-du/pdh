package cmd

import (
	"github.com/duyunis/pdh/cmd/receive"
	"github.com/duyunis/pdh/cmd/relay"
	"github.com/duyunis/pdh/cmd/send"
	"github.com/duyunis/pdh/cmd/version"
	"github.com/spf13/cobra"
)

// RootCmd describes the strange root command
var RootCmd = &cobra.Command{
	Use:              "pdh [sub]",
	Short:            "pdh",
	SilenceUsage:     true,
	PersistentPreRun: func(c *cobra.Command, args []string) {},
}

func init() {
	RootCmd.AddCommand(send.Cmd)
	RootCmd.AddCommand(receive.Cmd)
	RootCmd.AddCommand(relay.Cmd)
	RootCmd.AddCommand(version.Cmd)
}
