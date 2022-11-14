package sender

import (
	"fmt"
	"github.com/duyunzhi/discovery"
	"github.com/duyunzhi/pdh/common"
	"github.com/duyunzhi/pdh/compress"
	"github.com/duyunzhi/pdh/files"
	"github.com/duyunzhi/pdh/message"
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/proto"
	"github.com/duyunzhi/pdh/tools"
	"github.com/duyunzhi/pdh/transmit"
	"github.com/duyunzhi/pdh/transmit/client"
	"github.com/duyunzhi/pdh/transmit/server"
	"github.com/duyunzhi/progress_bar"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Sender struct {
	TotalFilesSize            int64
	longestFilename           int
	TotalNumberOfContents     int
	FilesToTransferCurrentNum int

	fs            *files.Files
	gs            *server.GrpcServer
	gc            *client.GrpcClient
	opt           *options.SenderOptions
	fileHandleMsg chan *proto.Message
	quit          chan bool
	done          chan bool
}

// Send files
func (s *Sender) Send(filePaths []string) {
	fs, err := files.GetFilesInfo(filePaths, s.opt.Zip)
	if err != nil {
		tools.Println(tools.Red, fmt.Sprintf("get files info error: %s", err))
		os.Exit(1)
	}
	s.fs = fs

	s.TotalNumberOfContents = len(s.fs.FilesInfo)

	if s.opt.LocalNetwork {
		err = s.sendWithLocalNetwork()
	} else {
		err = s.sendWithLocalNetwork()
		err = s.sendWithRelay()
	}
	if err != nil {
		tools.Println(tools.Red, fmt.Sprintf("send files error: %s", err))
		os.Exit(1)
	}
	go s.signal()
	<-s.done
}

func (s *Sender) sendWithLocalNetwork() error {
	opt := &discovery.Options{
		Duration:       -1,
		BroadcastDelay: time.Second,
		Payload:        []byte(s.opt.ShareCode),
	}

	broadcast := discovery.NewBroadcast(opt)
	broadcast.StartAsSync()

	grpcServer := server.NewPdhGrpcServer(&options.GrpcServerOptions{Address: "0.0.0.0", Ports: s.opt.LocalPort})
	grpcServer.AddHandler(s)
	go grpcServer.Start()
	return nil
}

func (s *Sender) sendWithRelay() error {
	gc := client.NewPdhGrpcClient(s.opt.Relay)
	gc.AddHandler(s)
	err := gc.Start()
	if err != nil {
		return err
	}
	s.gc = gc
	err = s.gc.Send(message.NewMessage(proto.MessageType_CreateChannel, []byte(s.opt.ShareCode)))
	return err
}

func (s *Sender) Done() {
	if s.opt.Zip {
		// delete zip file
		for _, info := range s.fs.FilesInfo {
			if strings.HasSuffix(info.Name, ".zip") {
				_ = os.Remove(info.Name)
			}
		}
	}
	// sleep, send an end message to the other.
	time.Sleep(time.Second)
	s.done <- true
}

func (s *Sender) signal() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-interrupt:
			if s.gc != nil {
				_ = s.gc.Send(message.NewMessage(proto.MessageType_Interrupt, nil))
			}
			s.Done()
		}
	}
}

func (s *Sender) HandleMessage(stream transmit.GrpcStream, msg *proto.Message) {
	var err error
	switch msg.MessageType {
	case proto.MessageType_LocalNetworkMode:
		// local network mode, stop relay client
		s.gc.Stop()
		s.gc = nil
	case proto.MessageType_Interrupt:
		fmt.Println("send interrupt...")
		s.Done()
	case proto.MessageType_Cancel:
		fmt.Println("send cancel")
		s.Done()
	case proto.MessageType_CreateChannelSuccess:
		fmt.Println("channel created")
		err := s.sendCollectFiles()
		if err != nil {
			fmt.Println("collect files error: ", err)
			os.Exit(1)
		}
		fmt.Println("share code is:", s.opt.ShareCode)
		fmt.Println("on the other computer run")
		fmt.Println()
		if s.opt.LocalNetwork {
			fmt.Println("pdh receive --local", s.opt.ShareCode)
		} else if s.opt.Relay == common.PublicRelay {
			if s.opt.LocalNetwork && s.opt.LocalPort != common.DefaultLocalPort {
				fmt.Println("pdh receive --local --local-port", s.opt.LocalPort, s.opt.ShareCode)
			} else {
				fmt.Println("pdh receive", s.opt.ShareCode)
			}
		} else {
			fmt.Println("pdh receive --relay", s.opt.Relay, s.opt.ShareCode)
		}
	case proto.MessageType_CreateChannelFailed:
		tools.Println(tools.Red, "create channel failed.")
		s.Done()
	case proto.MessageType_GetFileStat:
		fileStat := &message.FileStatPayload{
			FilesSize:    s.TotalFilesSize,
			FilesNumber:  int64(len(s.fs.FilesInfo)),
			FolderNumber: int64(s.fs.TotalNumberFolders),
		}
		payload, _ := fileStat.Bytes(message.JSONProtocol)
		err = stream.Send(message.NewMessage(proto.MessageType_FileStat, payload))
		if err != nil {
			tools.Println(tools.Red, "stream is error.")
			s.Done()
		}
	case proto.MessageType_RefuseReceive:
		tools.Println(tools.Yellow, "the other refused receive.")
		s.Done()
	case proto.MessageType_SkipFile, proto.MessageType_ReadyForReceive, proto.MessageType_FileFinish:
		s.fileHandleMsg <- msg
	case proto.MessageType_AgreeReceive:
		// start send files
		fmt.Println()
		fmt.Println("Sending...")
		fmt.Println()
	LOOP:
		for index, fileInfo := range s.fs.FilesInfo {
			fileInfoPayload := &message.FileInfoPayload{
				FileInfo: fileInfo,
			}
			payload, _ := fileInfoPayload.Bytes(message.JSONProtocol)
			err = stream.Send(message.NewMessage(proto.MessageType_FileInfo, payload))
			if err != nil {
				tools.Println(tools.Red, "stream is error.")
				s.Done()
				return
			}

		HANDLE:
			for {
				select {
				case m := <-s.fileHandleMsg:
					switch m.MessageType {
					case proto.MessageType_SkipFile:
						if index == len(s.fs.FilesInfo)-1 {
							// no file to send
							_ = stream.Send(message.NewMessage(proto.MessageType_FileFinish, nil))
						}
						continue LOOP
					case proto.MessageType_ReadyForReceive:
						break HANDLE
					}
				}
			}

			barOpt := &progress_bar.Options{
				Describe:     fileInfo.Name,
				Graph:        ">",
				IsBytes:      true,
				ShowPercent:  true,
				ShowDuration: true,
			}
			bar := progress_bar.NewBarWithOptions(fileInfo.Size, barOpt)
			filePath := path.Join(fileInfo.FolderSource, fileInfo.Name)
			reading, err := os.Open(filePath)
			if err != nil {
				return
			}
			readingPosition := int64(0)

			finish := false
			EOF := false

			for {
				data := make([]byte, common.MaxBufferSize/2)
				n, err := reading.ReadAt(data, readingPosition)
				if err != nil {
					if err == io.EOF {
						finish = true
						EOF = true
					} else {
						tools.Println(tools.Red, fmt.Sprintf("read file error: %s", err))
						s.Done()
						return
					}
				}
				dataToSend := compress.Compress(data[:n])
				readingPosition += int64(n)
				pl := &message.FileDataPayload{
					Data:     dataToSend,
					Position: readingPosition,
					EOF:      EOF,
				}
				filePayload, _ := pl.Bytes(message.JSONProtocol)
				fdm := message.NewMessage(proto.MessageType_FileData, filePayload)
				err = stream.Send(fdm)
				if err != nil {
					tools.Println(tools.Red, "stream is error.")
					s.Done()
					return
				}
				bar.Add(readingPosition)
				if finish {
					bar.Finish()
					break
				}
			}
		}
		fmt.Println("Send Completed!")
		s.Done()
	}
}

func (s *Sender) sendCollectFiles() (err error) {
	for i, fileInfo := range s.fs.FilesInfo {
		var fullPath string
		fullPath = fileInfo.FolderSource + string(os.PathSeparator) + fileInfo.Name
		fullPath = filepath.Clean(fullPath)

		if len(fileInfo.Name) > s.longestFilename {
			s.longestFilename = len(fileInfo.Name)
		}

		if os.FileMode(fileInfo.Mode)&os.ModeSymlink != 0 {
			fileInfo.Symlink, err = os.Readlink(fullPath)
			if err != nil {
			}
		}
		s.TotalFilesSize += fileInfo.Size
		if err != nil {
			return
		}
		fmt.Printf("\r                                 ")
		fmt.Printf("\rSending %d files (%s)", i, tools.ByteCountDecimal(s.TotalFilesSize))
	}
	fileName := fmt.Sprintf("%d files", len(s.fs.FilesInfo))
	folderName := fmt.Sprintf("%d folders", s.fs.TotalNumberFolders)
	if len(s.fs.FilesInfo) == 1 {
		fileName = fmt.Sprintf("'%s'", s.fs.FilesInfo[0].Name)
	}

	fmt.Printf("\r                                 ")
	if s.fs.TotalNumberFolders > 0 {
		fmt.Printf("\rSending %s and %s (%s)\n", fileName, folderName, tools.ByteCountDecimal(s.TotalFilesSize))
	} else {
		fmt.Printf("\rSending %s (%s)\n", fileName, tools.ByteCountDecimal(s.TotalFilesSize))
	}
	return
}

func checkOptions(opt *options.SenderOptions) {
	if tools.IsBlank(opt.ShareCode) {
		tools.Println(tools.Red, "share code can't empty")
		os.Exit(1)
	}
	if opt.LocalNetwork && opt.Relay != common.PublicRelay {
		tools.Println(tools.Yellow, "you enable local network, relay will disabled")
	}
}

func NewSender(opt *options.SenderOptions) *Sender {
	checkOptions(opt)
	return &Sender{
		opt:           opt,
		fileHandleMsg: make(chan *proto.Message, 10),
		quit:          make(chan bool, 1),
		done:          make(chan bool, 1),
	}
}
