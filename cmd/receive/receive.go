package receive

import (
	"fmt"
	"github.com/duyunzhi/pdh/common"
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/receiver"
	"github.com/spf13/cobra"
)

var opt = &options.ReceiverOptions{}

var Cmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive file(s), or folder (see options with pdh receive -h)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("no share code")
			return
		}
		opt.ShareCode = args[0]
		rec := receiver.NewReceiver(opt)
		rec.Receive()
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&opt.Relay, "relay", "", common.PublicRelay, "relay address")
	Cmd.PersistentFlags().StringVarP(&opt.OutPath, "out", "o", "", "receive path")
	Cmd.PersistentFlags().BoolVarP(&opt.LocalNetwork, "local", "", false, "use local network (default: false)")
	Cmd.PersistentFlags().StringVarP(&opt.LocalPort, "local-port", "", "6880", "effect when the local network is enabled")
}
