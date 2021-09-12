package sockets

import (
	"net"
	"sync"
)

var bytesPool = sync.Pool{
	New: func() interface{} { return []byte{} },
}

type Sockets struct{}

var socketsManager Sockets

func GetSocketsManager() *Sockets {
	return &socketsManager
}

func (s *Sockets) GetHttpConnection(ip string, isSecure bool) (net.Conn, error) {
	if isSecure {
		return getHttpsConnection(ip)
	} else {
		return getHttpConnection(ip)
	}
}
