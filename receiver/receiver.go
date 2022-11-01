package receiver

import (
	"fmt"
	"github.com/duyunzhi/pdh/common"
	"github.com/duyunzhi/pdh/message"
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/proto"
	"github.com/duyunzhi/pdh/tools"
	"github.com/duyunzhi/pdh/transmit/client"
	"os"
	"strings"
	"time"
)

type Receiver struct {
	opt       *options.ReceiverOptions
	gc        *client.GrpcClient
	filesSize int64
	done      chan bool
}

func (r *Receiver) Receive() {
	fmt.Print("Starting receive...")
	fmt.Print("\r                              ")
	var err error
	if r.opt.LocalNetwork {
		err = r.receiveFromLocalNetwork()
	} else {
		err = r.receiveFromRelay()
	}
	if err != nil {
		tools.Println(tools.Red, fmt.Sprintf("error occurred, %s", err))
		os.Exit(1)
	}
	<-r.done
}

func (r *Receiver) receiveFromLocalNetwork() error {
	return nil
}

func (r *Receiver) receiveFromRelay() error {
	fmt.Print("\rConnecting...")
	gc := client.NewPdhGrpcClient(r.opt.Relay)
	gc.AddHandler(r)
	err := gc.Start()
	if err != nil {
		return err
	}
	r.gc = gc
	return r.gc.Send(message.NewMessage(proto.MessageType_JoinChannel, []byte(r.opt.ShareCode)))
}

func (r *Receiver) Done() {
	// sleep, send an end message to the other.
	time.Sleep(time.Second)
	r.done <- true
}

func (r *Receiver) HandleMessage(stream proto.PdhService_TransmitClient, msg *proto.Message) {
	var err error
	switch msg.MessageType {
	case proto.MessageType_ChannelFull:
		tools.Println(tools.Red, "channel is full, someone else has already received it.")
		os.Exit(1)
	case proto.MessageType_JoinChannelSuccess:
		fmt.Print("\rjoin channel success.")
		err = stream.Send(message.NewMessage(proto.MessageType_GetFileInfo, nil))
		if err != nil {
			tools.Println(tools.Red, "stream is error.")
			os.Exit(1)
		}
	case proto.MessageType_ChannelNotFound:
		tools.Println(tools.Red, "channel not found, please check your share code.")
		os.Exit(1)
	case proto.MessageType_JoinChannelFailed:
		tools.Println(tools.Red, "join channel failed.")
		os.Exit(1)
	case proto.MessageType_FileInfo:
		pm, err := message.ParseMessagePayload(msg)
		if err != nil {
			tools.Println(tools.Red, "get file info failed.")
			os.Exit(1)
		}
		stat := pm.(*message.FileStatPayload)
		fmt.Printf("\rAccept %d files and %d folders (%s)? (Y/n)", stat.FilesNumber, stat.FolderNumber, tools.ByteCountDecimal(stat.FilesSize))
		fmt.Println()
		choice := strings.ToLower(tools.GetInput(""))
		if choice != "" && choice != "y" && choice != "yes" {
			_ = stream.Send(message.NewMessage(proto.MessageType_RefuseReceive, nil))
			r.Done()
			return
		}
		r.filesSize = stat.FilesSize
		err = stream.Send(message.NewMessage(proto.MessageType_AgreeReceive, nil))
		if err != nil {
			tools.Println(tools.Red, "stream is error.")
			os.Exit(1)
		}
	case proto.MessageType_CreateChannel:
		fmt.Println("start receive")
	}

}

func checkOptions(opt *options.ReceiverOptions) {
	if tools.IsBlank(opt.ShareCode) {
		tools.Println(tools.Red, "share code can't empty")
		os.Exit(1)
	}
	if !opt.LocalNetwork && tools.IsBlank(opt.Relay) {
		tools.Println(tools.Red, "relay address can't empty")
		os.Exit(1)
	}
	if opt.LocalNetwork && opt.Relay != common.PublicRelay {
		tools.Println(tools.Yellow, "you enable local network, relay will disabled")
	}
}

func NewReceiver(opt *options.ReceiverOptions) *Receiver {
	checkOptions(opt)
	return &Receiver{
		opt:  opt,
		done: make(chan bool, 1),
	}
}
