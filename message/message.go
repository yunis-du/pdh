package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duyunzhi/pdh/files"
	"github.com/duyunzhi/pdh/proto"
)

// Protocol used to transport messages
type Protocol string

const (
	// RawProtocol is raw messages
	RawProtocol Protocol = "raw"
	// JSONProtocol is used for JSON encoded messages
	JSONProtocol Protocol = "json"
)

// Message is the interface of a message to send over the wire
type Message interface {
	Bytes(protocol Protocol) ([]byte, error)
}

type ShareCodePayload struct {
	ShareCode string
}

func (c *ShareCodePayload) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == JSONProtocol {
		return json.Marshal(c)
	}
	return nil, nil
}

type FileStatPayload struct {
	FilesSize    int64 `json:"FilesSize,omitempty"`
	FilesNumber  int64 `json:"FilesNumber,omitempty"`
	FolderNumber int64 `json:"FolderNumber,omitempty"`
}

func (f *FileStatPayload) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == JSONProtocol {
		return json.Marshal(f)
	}
	return nil, nil
}

type FileInfoPayload struct {
	FileInfo *files.FileInfo
}

func (f *FileInfoPayload) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == JSONProtocol {
		return json.Marshal(f)
	}
	return nil, nil
}

type FileDataPayload struct {
	Data     []byte
	Position int64
	EOF      bool
}

func (f *FileDataPayload) Bytes(protocol Protocol) ([]byte, error) {
	if protocol == JSONProtocol {
		return json.Marshal(f)
	}
	return nil, nil
}

func ParseMessagePayload(msg *proto.Message) (Message, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}
	payload := msg.Payload
	switch msg.MessageType {
	case proto.MessageType_CreateChannel, proto.MessageType_JoinChannel:
		shareCode := ""
		if payload != nil {
			shareCode = string(payload)
		}
		return &ShareCodePayload{ShareCode: shareCode}, nil
	case proto.MessageType_FileStat:
		if payload != nil {
			var fs FileStatPayload
			err := json.Unmarshal(payload, &fs)
			if err != nil {
				return nil, err
			}
			return &fs, nil
		}
	case proto.MessageType_FileInfo:
		if payload != nil {
			var fi FileInfoPayload
			err := json.Unmarshal(payload, &fi)
			if err != nil {
				return nil, err
			}
			return &fi, nil
		}
	case proto.MessageType_FileData:
		if payload != nil {
			var fd FileDataPayload
			err := json.Unmarshal(payload, &fd)
			if err != nil {
				return nil, err
			}
			return &fd, nil
		}
	default:
		return nil, errors.New(fmt.Sprintf("unknown message type: [%s]", msg.MessageType.String()))
	}
	return nil, nil
}

func NewMessage(messageType proto.MessageType, payload []byte) *proto.Message {
	return &proto.Message{
		MessageType: messageType,
		Payload:     payload,
	}
}
