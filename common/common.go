package common

import "sync"

// RWMutex is a wrapper around sync.RWMutex
type RWMutex struct {
	sync.RWMutex
}

const (
	PublicRelay   = "127.0.0.1:50051"
	MaxBufferSize = 1024 * 64
)
