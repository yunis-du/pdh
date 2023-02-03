package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/duyunis/pdh/common"
	"github.com/duyunis/pdh/proto"
	"github.com/duyunis/pdh/transmit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	common.RWMutex
	conn     *grpc.ClientConn
	client   proto.PdhServiceClient
	stream   proto.PdhService_TransmitClient
	handlers []transmit.MessageHandler
	target   string
}

func (p *GrpcClient) Start() error {
	conn, err := grpc.Dial(p.target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return err
	}
	p.conn = conn
	p.client = proto.NewPdhServiceClient(conn)
	stream, err := p.client.Transmit(context.Background())
	if err != nil {
		return err
	}
	p.stream = stream

	go p.receive()
	return nil
}

func (p *GrpcClient) Stop() {
	if p.stream != nil {
		_ = p.stream.CloseSend()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

// AddHandler add a message handler
func (p *GrpcClient) AddHandler(handler transmit.MessageHandler) {
	p.Lock()
	defer p.Unlock()
	p.handlers = append(p.handlers, handler)
}

// RemoveHandler remove a message handler
func (p *GrpcClient) RemoveHandler(handler transmit.MessageHandler) {
	p.Lock()
	defer p.Unlock()

	for i, h := range p.handlers {
		if handler == h {
			p.handlers = append(p.handlers[:i], p.handlers[i+1:]...)
			break
		}
	}
}

func (p *GrpcClient) Send(msg *proto.Message) error {
	if p.stream == nil || p.client == nil {
		p.Start()
	}
	if p.stream != nil {
		err := p.stream.Send(msg)
		return err
	}
	return errors.New("stream is nil")
}

// ReceiveAsBlock will be blocked
func (p *GrpcClient) receive() {
	for {
		msg, err := p.stream.Recv()
		if err != nil {
			return
		}
		streamWrapper := transmit.NewClientStreamWrapper(p.stream)
		go p.dispatchMessage(msg, streamWrapper)
	}
}

func (p *GrpcClient) dispatchMessage(msg *proto.Message, cw *transmit.ClientStreamWrapper) {
	for _, handler := range p.handlers {
		handler.HandleMessage(cw, msg)
	}
}

func NewPdhGrpcClient(target string) *GrpcClient {
	return &GrpcClient{
		target:   target,
		handlers: make([]transmit.MessageHandler, 0),
	}
}
