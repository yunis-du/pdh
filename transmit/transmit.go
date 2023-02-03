package transmit

import (
	"github.com/duyunis/pdh/proto"
	"sync/atomic"
)

type MessageHandler interface {
	HandleMessage(stream GrpcStream, msg *proto.Message)
}

type GrpcStream interface {
	Send(msg *proto.Message) error
}

type ServerStreamWrapper struct {
	Stream    proto.PdhService_TransmitServer
	handlers  []MessageHandler
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

func NewServerStreamWrapper(stream proto.PdhService_TransmitServer) *ServerStreamWrapper {
	return &ServerStreamWrapper{
		Stream: stream,
		Ch:     make(chan *proto.Message, 10),
	}
}

type ClientStreamWrapper struct {
	Stream proto.PdhService_TransmitClient
}

func (c *ClientStreamWrapper) Send(msg *proto.Message) error {
	return c.Stream.Send(msg)
}

func NewClientStreamWrapper(stream proto.PdhService_TransmitClient) *ClientStreamWrapper {
	return &ClientStreamWrapper{
		Stream: stream,
	}
}
