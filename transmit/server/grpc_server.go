package server

import (
	"errors"
	"github.com/duyunzhi/pdh/common"
	"github.com/duyunzhi/pdh/options"
	"github.com/duyunzhi/pdh/proto"
	"github.com/duyunzhi/pdh/tools"
	"github.com/duyunzhi/pdh/transmit"
	"google.golang.org/grpc"
	"net"
)

type GrpcServer struct {
	common.RWMutex
	proto.UnimplementedPdhServiceServer
	options  *options.GrpcServerOptions
	handlers []transmit.ServerMessageHandler
	streams  map[string]*transmit.ServerStreamWrapper
	server   *grpc.Server
}

func (p *GrpcServer) Transmit(stream proto.PdhService_TransmitServer) error {
	p.Lock()
	genKey := tools.GenRandStr(8, ".")
	sw := transmit.NewServerStreamWrapper(genKey, stream)
	p.streams[genKey] = sw
	p.Unlock()
	for {
		msg, err := stream.Recv()
		if err != nil {
			p.Lock()
			delete(p.streams, genKey)
			p.Unlock()
			return err
		}
		p.dispatchMessage(msg, sw)
	}
}

// AddHandler add a message handler
func (p *GrpcServer) AddHandler(handler transmit.ServerMessageHandler) {
	p.Lock()
	defer p.Unlock()
	p.handlers = append(p.handlers, handler)
}

// RemoveHandler remove a message handler
func (p *GrpcServer) RemoveHandler(handler transmit.ServerMessageHandler) {
	p.Lock()
	defer p.Unlock()

	for i, h := range p.handlers {
		if handler == h {
			p.handlers = append(p.handlers[:i], p.handlers[i+1:]...)
			break
		}
	}
}

func (p *GrpcServer) Send(stream proto.PdhService_TransmitServer, message *proto.Message) error {
	if stream != nil {
		err := stream.Send(message)
		if err != nil {
			return err
		}
	}
	return errors.New("send error stream is nil")
}

func (p *GrpcServer) dispatchMessage(msg *proto.Message, sw *transmit.ServerStreamWrapper) {
	if sw.WriteToCh.Load() {
		sw.Ch <- msg
	} else {
		for _, handler := range p.handlers {
			handler.HandleMessage(sw, msg)
		}
	}
}

func (p *GrpcServer) Start() error {
	network := "tcp"
	hostPort := net.JoinHostPort(p.options.Address, p.options.Ports)
	listen, err := net.Listen(network, hostPort)
	if err != nil {
		return err
	}
	proto.RegisterPdhServiceServer(p.server, p)
	err = p.server.Serve(listen)
	if err != nil {
		return err
	}
	return nil
}

func (p *GrpcServer) Stop() {
	p.server.Stop()
}

func NewPdhGrpcServer(opt *options.GrpcServerOptions) *GrpcServer {
	return &GrpcServer{
		server:   grpc.NewServer(),
		options:  opt,
		handlers: make([]transmit.ServerMessageHandler, 0),
		streams:  make(map[string]*transmit.ServerStreamWrapper, 0),
	}
}
