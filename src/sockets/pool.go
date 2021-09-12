package sockets

import "sync"

var bytesPool = sync.Pool{
	New: func() interface{} { return []byte{} },
}
