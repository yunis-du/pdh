package receiver

import (
	"errors"
	"fmt"
	"github.com/duyunzhi/discovery"
	"github.com/duyunzhi/pdh/common"
	"github.com/duyunzhi/pdh/compress"
	"github.com/duyunzhi/pdh/message"
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/proto"
	"github.com/duyunzhi/pdh/tools"
	"github.com/duyunzhi/pdh/transmit"
	"github.com/duyunzhi/pdh/transmit/client"
	"github.com/duyunzhi/progress_bar"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Receiver struct {
	opt                 *options.ReceiverOptions
	gc                  *client.GrpcClient
	wg                  sync.WaitGroup
	currentFile         *os.File
	currentFinish       bool
	filesSize           int64
	fileHandleMsg       chan *proto.Message
	done                chan bool
	latestFileWriteDone chan bool
}

func (r *Receiver) Receive() {
	var err error

	err = r.receiveFromLocalNetwork()
	if err != nil {
		err = r.receiveFromRelay()
	}
	if err != nil {
		tools.Println(tools.Red, fmt.Sprintf("error occurred, %s", err))
		r.Done()
	}
	go r.signal()
	<-r.done
}

func (r *Receiver) receiveFromLocalNetwork() error {
	opt := &discovery.Options{
		Limit:     1,
		TimeLimit: time.Second * 5,
		Payload:   []byte(r.opt.ShareCode),
	}
	discover := discovery.NewDiscover(opt)
	broadcast, err := discover.DiscoverBroadcast()
	if err != nil {
		return err
	}
	if len(broadcast) > 0 {
		hostPort := net.JoinHostPort(broadcast[0].Address, r.opt.LocalPort)
		gc := client.NewPdhGrpcClient(hostPort)
		err = gc.Start()

		if err != nil {
			tools.Println(tools.Red, err)
			r.Done()
		}
		err = gc.Send(message.NewMessage(proto.MessageType_LocalNetworkMode, nil))
		err = gc.Send(message.NewMessage(proto.MessageType_GetFileStat, nil))
		if err != nil {
			tools.Println(tools.Red, "stream is error.")
			r.Done()
		}
		r.gc = gc
		gc.AddHandler(r)
		return nil
	}
	return errors.New("not discovery on local network")
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
	r.latestFileWriteDone <- true
}

func (r *Receiver) signal() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-interrupt:
			_ = r.gc.Send(message.NewMessage(proto.MessageType_Interrupt, nil))
			r.Done()
		}
	}
}

func (r *Receiver) HandleMessage(stream transmit.GrpcStream, msg *proto.Message) {
	var err error
	switch msg.MessageType {
	case proto.MessageType_Interrupt:
		fmt.Println("\rreceive interrupt...")
		r.Done()
	case proto.MessageType_ChannelFull:
		tools.Println(tools.Red, "\rchannel is full, someone else has already received it.")
		r.Done()
	case proto.MessageType_JoinChannelSuccess:
		fmt.Print("\rjoin channel success.")
		err = stream.Send(message.NewMessage(proto.MessageType_GetFileStat, nil))
		if err != nil {
			tools.Println(tools.Red, "stream is error.")
			r.Done()
		}
	case proto.MessageType_ChannelNotFound:
		tools.Println(tools.Red, "\rchannel not found, please check your share code.")
		r.Done()
	case proto.MessageType_JoinChannelFailed:
		tools.Println(tools.Red, "\rjoin channel failed.")
		r.Done()
	case proto.MessageType_FileFinish:
		r.Done()
	case proto.MessageType_FileStat:
		pm, err := message.ParseMessagePayload(msg)
		if err != nil {
			tools.Println(tools.Red, "\rget file stat failed.")
			r.Done()
			return
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
			tools.Println(tools.Red, "\rstream is error.")
			r.Done()
		}
		fmt.Println()
		fmt.Println("Receiving...")
		fmt.Println()
		r.wg.Add(int(stat.FilesNumber))
		go func() {
			r.wg.Wait()
			fmt.Println("Receive Completed!")
			r.Done()
		}()
	case proto.MessageType_FileData:
		r.fileHandleMsg <- msg
	case proto.MessageType_FileInfo:
		if r.currentFile != nil {
			<-r.latestFileWriteDone
		}
		pm, err := message.ParseMessagePayload(msg)
		if err != nil {
			tools.Println(tools.Red, fmt.Sprintf("get file info failed: %s", err))
			r.Done()
			return
		}
		infoPayload := pm.(*message.FileInfoPayload)
		if infoPayload == nil {
			tools.Println(tools.Red, fmt.Sprintf("get file info failed: %s", err))
			r.Done()
			return
		}
		fileInfo := infoPayload.FileInfo
		if fileInfo != nil {
			pathToDir := path.Join(r.opt.OutPath, fileInfo.FolderRemote)
			pathToFile := path.Join(r.opt.OutPath, fileInfo.FolderRemote, fileInfo.Name)
			boo := tools.IsFile(pathToDir)
			if !boo {
				if err = os.MkdirAll(pathToDir, os.ModePerm); err != nil {
					tools.Println(tools.Red, fmt.Sprintf("create folder failed, %s", err))
					return
				}
			}
			boo = tools.IsFile(pathToFile)
			if boo {
				// file existed
				fmt.Printf("\rFile %s is existed, do you want to overwrite it? (Y/n)", fileInfo.Name)
				fmt.Println()
				choice := strings.ToLower(tools.GetInput(""))
				if choice != "" && choice != "y" && choice != "yes" {
					_ = stream.Send(message.NewMessage(proto.MessageType_SkipFile, nil))
					return
				}
				r.currentFile, err = os.OpenFile(pathToFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
			} else {
				r.currentFile, err = os.Create(pathToFile)
			}
			if err != nil {
				tools.Println(tools.Red, fmt.Sprintf("create or open file [%s] failed, %s", pathToFile, err))
				r.Done()
				return
			}
			err = r.currentFile.Truncate(fileInfo.Size)
			if err != nil {
				tools.Println(tools.Red, fmt.Sprintf("could not truncate [%s]: %s", pathToFile, err))
				r.Done()
				return
			}
			// ready
			err = stream.Send(message.NewMessage(proto.MessageType_ReadyForReceive, nil))
			if err != nil {
				tools.Println(tools.Red, fmt.Sprintf("stream is error: %s", err))
				r.Done()
				return
			}
			barOpt := &progress_bar.Options{
				Describe:     fileInfo.Name,
				Graph:        ">",
				IsBytes:      true,
				ShowPercent:  true,
				ShowDuration: true,
			}
			bar := progress_bar.NewBarWithOptions(fileInfo.Size, barOpt)

			writePosition := int64(0)
			go func() {
			LOOP:
				for {
					select {
					case m := <-r.fileHandleMsg:
						switch m.MessageType {
						case proto.MessageType_FileData:
							pmp, _ := message.ParseMessagePayload(m)
							fileDataMsg := pmp.(*message.FileDataPayload)
							if fileDataMsg.Data != nil {
								receiveData := compress.Decompress(fileDataMsg.Data)
								_, err = r.currentFile.Write(receiveData)
								if err != nil {
									tools.Println(tools.Red, fmt.Sprintf("write file [%s] failed: %s", pathToFile, err))
									_ = os.Remove(r.currentFile.Name())
									r.Done()
									break LOOP
								}
								writePosition = fileDataMsg.Position
								bar.Add(writePosition)
								if fileDataMsg.EOF {
									bar.Finish()
									err = r.currentFile.Close()
									r.latestFileWriteDone <- true
									break LOOP
								}
							}
						}
					}
				}
				r.wg.Done()
			}()
		}
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
		opt:                 opt,
		fileHandleMsg:       make(chan *proto.Message, 10),
		done:                make(chan bool, 1),
		latestFileWriteDone: make(chan bool, 1),
	}
}
