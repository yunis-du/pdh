package common

import "sync"

// RWMutex is a wrapper around sync.RWMutex
type RWMutex struct {
	sync.RWMutex
}

const (
	PublicRelay = "101.35.47.168:6880"
	//PublicRelay      = "127.0.0.1:50051"
	DefaultLocalPort = "6880"
	MaxBufferSize    = 1024 * 64
)
