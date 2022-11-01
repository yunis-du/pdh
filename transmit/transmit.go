package transmit

import (
	"github.com/duyunzhi/pdh/proto"
	"sync/atomic"
)

type ServerMessageHandler interface {
	HandleMessage(sw *ServerStreamWrapper, msg *proto.Message)
}

type ClientMessageHandler interface {
	HandleMessage(stream proto.PdhService_TransmitClient, msg *proto.Message)
}

type ServerStreamWrapper struct {
	Key       string
	Stream    proto.PdhService_TransmitServer
	handlers  []ServerMessageHandler
	Ch        chan *proto.Message
	WriteToCh atomic.Bool
}

func (s *ServerStreamWrapper) Send(msg *proto.Message) error {
	return s.Stream.Send(msg)
}

func (s *ServerStreamWrapper) StartWriteToChannel() {
	s.WriteToCh.Store(true)
}

func (s *ServerStreamWrapper) StopWriteToChannel() {
	s.WriteToCh.Store(false)
}

func NewServerStreamWrapper(key string, stream proto.PdhService_TransmitServer) *ServerStreamWrapper {
	return &ServerStreamWrapper{
		Key:    key,
		Stream: stream,
		Ch:     make(chan *proto.Message, 10),
	}
}
