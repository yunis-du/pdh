package sender

import (
	"fmt"
	"github.com/duyunzhi/pdh/common"
	"github.com/duyunzhi/pdh/files"
	"github.com/duyunzhi/pdh/message"
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/proto"
	"github.com/duyunzhi/pdh/tools"
	"github.com/duyunzhi/pdh/transmit/client"
	"github.com/duyunzhi/pdh/transmit/server"
	"os"
	"path/filepath"
	"time"
)

type Sender struct {
	TotalFilesSize            int64
	longestFilename           int
	TotalNumberOfContents     int
	FilesToTransferCurrentNum int

	fs   *files.Files
	gs   *server.GrpcServer
	gc   *client.GrpcClient
	opt  *options.SenderOptions
	quit chan bool
	done chan bool
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
		err = s.sendWithRelay()
	}
	if err != nil {
		tools.Println(tools.Red, fmt.Sprintf("send files error: %s", err))
		os.Exit(1)
	}
	<-s.done
}

func (s *Sender) sendWithLocalNetwork() error {
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
	// sleep, send an end message to the other.
	time.Sleep(time.Second)
	s.done <- true
}

func (s *Sender) HandleMessage(stream proto.PdhService_TransmitClient, msg *proto.Message) {
	var err error
	switch msg.MessageType {
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
			fmt.Println("pdh receive", s.opt.ShareCode)
		} else {
			fmt.Println("pdh receive --relay", s.opt.Relay, s.opt.ShareCode)
		}
	case proto.MessageType_CreateChannelFailed:
		tools.Println(tools.Red, "create channel failed.")
		os.Exit(1)
	case proto.MessageType_GetFileInfo:
		fileStat := &message.FileStatPayload{
			FilesSize:    s.TotalFilesSize,
			FilesNumber:  int64(len(s.fs.FilesInfo)),
			FolderNumber: int64(s.fs.TotalNumberFolders),
		}
		payload, _ := fileStat.Bytes(message.JSONProtocol)
		err = stream.Send(message.NewMessage(proto.MessageType_FileInfo, payload))
		if err != nil {
			tools.Println(tools.Red, "stream is error.")
			os.Exit(1)
		}
	case proto.MessageType_RefuseReceive:
		tools.Println(tools.Yellow, "the other refused receive.")
		s.Done()
	case proto.MessageType_AgreeReceive:
		// start send files
		/*for _, fileInfo := range s.fs.FilesInfo {

		}*/
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

		if s.opt.HashAlgorithm == "" {
			s.opt.HashAlgorithm = "xxhash"
		}

		fileInfo.Hash, err = tools.HashFile(fullPath, s.opt.HashAlgorithm)
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
		opt:  opt,
		quit: make(chan bool, 1),
		done: make(chan bool, 1),
	}
}
