package relay

import (
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/relay"
	"github.com/spf13/cobra"
	"log"
)

var opt = &options.RelayOptions{}

var Cmd = &cobra.Command{
	Use:   "relay",
	Short: "Start your own relay",
	Run: func(cmd *cobra.Command, args []string) {
		re := relay.NewRelay(opt)
		err := re.Run()
		if err != nil {
			log.Printf("start relay error: %s\n", err)
		}
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&opt.RelayHost, "host", "", "0.0.0.0", "relay host")
	Cmd.PersistentFlags().StringVarP(&opt.RelayPort, "port", "", "50051", "relay port")
}
