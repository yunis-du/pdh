package relay

import (
	"github.com/duyunis/pdh/common"
	"github.com/duyunis/pdh/message"
	"github.com/duyunis/pdh/options"
	"github.com/duyunis/pdh/proto"
	"github.com/duyunis/pdh/transmit"
	"github.com/duyunis/pdh/transmit/pipe"
	"github.com/duyunis/pdh/transmit/server"
	"log"
	"time"
)

type Relay struct {
	common.RWMutex
	options    *options.RelayOptions
	grpcServer *server.GrpcServer
	channels   map[string]*channel
}

type channel struct {
	owner     *transmit.ServerStreamWrapper
	visitor   *transmit.ServerStreamWrapper
	createdAt time.Time
	full      bool
	pipe      *pipe.Pipe
}

func (r *Relay) Run() error {
	go r.checkChannel()
	return r.grpcServer.Start()
}

func (r *Relay) RunAsync() {
	go r.checkChannel()
	go r.grpcServer.Start()
}

// checkChannel timeout and not connected will delete
func (r *Relay) checkChannel() {
	ticker := time.NewTicker(time.Second * 3)
	for {
		select {
		case <-ticker.C:
			for key := range r.channels {
				ch := r.channels[key]
				if ch == nil {
					delete(r.channels, key)
					continue
				}
				err := ch.owner.Send(&proto.Message{MessageType: proto.MessageType_Ping})
				if err != nil {
					if ch.pipe != nil {
						ch.pipe.Stop()
					}
					delete(r.channels, key)
					continue
				}
				if ch.visitor != nil {
					err = ch.visitor.Send(&proto.Message{MessageType: proto.MessageType_Ping})
					if err != nil {
						if ch.pipe != nil {
							ch.pipe.Stop()
						}
						delete(r.channels, key)
						continue
					}
				}
				if time.Since(ch.createdAt) > time.Minute*30 {
					ch.pipe.Stop()
					delete(r.channels, key)
					continue
				}
			}
		}
	}
}

func (r *Relay) HandleMessage(stream transmit.GrpcStream, msg *proto.Message) {
	r.Lock()
	defer r.Unlock()
	switch msg.MessageType {
	case proto.MessageType_CreateChannel:
		parseMsg, err := message.ParseMessagePayload(msg)
		if err != nil {
			log.Printf("parse message error: %s\n", err)
			return
		}
		channelMsg := parseMsg.(*message.ShareCodePayload)
		shareCode := channelMsg.ShareCode
		if len(shareCode) > 0 {
			_, ok := r.channels[shareCode]
			if ok {
				_ = stream.Send(message.NewMessage(proto.MessageType_CreateChannelFailed, nil))
			} else {
				r.channels[shareCode] = &channel{
					owner:     stream.(*transmit.ServerStreamWrapper),
					createdAt: time.Now(),
				}
				_ = stream.Send(message.NewMessage(proto.MessageType_CreateChannelSuccess, nil))
			}
		} else {
			_ = stream.Send(message.NewMessage(proto.MessageType_CreateChannelFailed, nil))
		}
	case proto.MessageType_JoinChannel:
		parseMsg, err := message.ParseMessagePayload(msg)
		if err != nil {
			log.Printf("parse message error: %s\n", err)
			return
		}
		channelMsg := parseMsg.(*message.ShareCodePayload)
		shareCode := channelMsg.ShareCode
		if len(shareCode) > 0 {
			ch, ok := r.channels[shareCode]
			if ok {
				if ch.full {
					_ = stream.Send(message.NewMessage(proto.MessageType_ChannelFull, nil))
				} else {
					ch.visitor = stream.(*transmit.ServerStreamWrapper)
					// create pipe
					ch.pipe = pipe.CreatePipe(ch.owner, ch.visitor)
					ch.pipe.Start()

					_ = stream.Send(message.NewMessage(proto.MessageType_JoinChannelSuccess, nil))
				}
			} else {
				_ = stream.Send(message.NewMessage(proto.MessageType_ChannelNotFound, nil))
			}
		} else {
			_ = stream.Send(message.NewMessage(proto.MessageType_JoinChannelFailed, nil))
		}

	}
}

func (r *Relay) SendMessage(stream proto.PdhService_TransmitServer, msg *proto.Message) error {
	return stream.Send(msg)
}

func NewRelay(opt *options.RelayOptions) *Relay {
	grpcServer := server.NewPdhGrpcServer(&options.GrpcServerOptions{Address: opt.RelayHost, Ports: opt.RelayPort})
	relay := &Relay{
		options:    opt,
		channels:   make(map[string]*channel, 0),
		grpcServer: grpcServer,
	}
	// add message handler
	grpcServer.AddHandler(relay)
	return relay
}
