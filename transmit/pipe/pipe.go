package pipe

import (
	"github.com/duyunzhi/pdh/message"
	"github.com/duyunzhi/pdh/proto"
	"github.com/duyunzhi/pdh/transmit"
	"log"
	"sync/atomic"
)

type Pipe struct {
	first   *transmit.ServerStreamWrapper
	second  *transmit.ServerStreamWrapper
	quit    chan bool
	running atomic.Bool
}

func (p *Pipe) chanFromStream(stream proto.PdhService_TransmitServer) chan *proto.Message {
	ch := make(chan *proto.Message, 10)
	go func() {
		for {
			if !p.running.Load() {
				break
			}
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("receive message error: %s\n", err)
				return
			}
			ch <- msg
		}
	}()
	return ch
}

func (p *Pipe) Start() {

	p.running.Store(true)
	go func() {
		p.first.StartWriteToChannel()
		p.second.StartWriteToChannel()
	LOOP:
		for {
			select {
			case <-p.quit:
				break LOOP
			case m1 := <-p.first.Ch:
				err := p.second.Send(m1)
				if err != nil {
					_ = p.first.Send(message.NewMessage(proto.MessageType_Failed, []byte("stream is closed")))
					p.Stop()
					log.Printf("send message error: %s\n", err)
				}
			case m2 := <-p.second.Ch:
				err := p.first.Send(m2)
				if err != nil {
					_ = p.second.Send(message.NewMessage(proto.MessageType_Failed, []byte("stream is closed")))
					p.Stop()
					log.Printf("send message error: %s\n", err)
				}
			}
		}
	}()
}

func (p *Pipe) Stop() {
	p.notifyBoth()
	p.quit <- true
	p.running.Store(false)
	p.first.StopWriteToChannel()
	p.second.StopWriteToChannel()
}

func (p *Pipe) notifyBoth() {
	_ = p.first.Send(message.NewMessage(proto.MessageType_Cancel, nil))
	_ = p.second.Send(message.NewMessage(proto.MessageType_Cancel, nil))
}

func CreatePipe(first, second *transmit.ServerStreamWrapper) *Pipe {
	return &Pipe{
		first:  first,
		second: second,
		quit:   make(chan bool, 1),
	}
}
