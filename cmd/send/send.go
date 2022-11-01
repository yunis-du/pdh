package send

import (
	"github.com/duyunzhi/pdh/common"
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/sender"
	"github.com/duyunzhi/pdh/tools"
	"github.com/spf13/cobra"
)

var opt = &options.SenderOptions{}

var Cmd = &cobra.Command{
	Use:   "send",
	Short: "Send file(s), or folder (see options with pdh send -h)",
	Run: func(cmd *cobra.Command, args []string) {
		files := tools.GetAbsolutePaths(args)
		opt.ShareCode = tools.GenRandStr(4, "-")
		fileSender := sender.NewSender(opt)
		fileSender.Send(files)
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&opt.ShareCode, "shareCode", "c", "", "code used to connect to relay")
	Cmd.PersistentFlags().BoolVarP(&opt.Zip, "zip", "", false, "zip folder before sending (default: false)")
	Cmd.PersistentFlags().StringVarP(&opt.Relay, "relay", "", common.PublicRelay, "relay address")
	Cmd.PersistentFlags().BoolVarP(&opt.LocalNetwork, "local", "", false, "use local network (default: false)")
}
