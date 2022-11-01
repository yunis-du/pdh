package version

import (
	"github.com/duyunzhi/pdh/version"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of pdh",
	Run: func(cmd *cobra.Command, args []string) {
		version.PrintVersion()
	},
}
