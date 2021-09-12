package sockets

import (
	"log"
	"net"
	"time"
)

func GetHttpConnection(ip string) (net.Conn, error) {
	d := net.Dialer{
		Timeout:  3 * time.Second,
		Deadline: time.Now().Add(time.Second * 5),
	}
	conn, err := d.Dial("tcp", ip+":80")
	if err != nil {
		return nil, err
	}

	bytes := bytesPool.Get().([]byte)
	_, err = conn.Read(bytes)
	if err != nil {
		_ = conn.Close()
		log.Println(err)
		return nil, err
	}
	bytes = bytes[:0]
	bytesPool.Put(bytes)

	return conn, nil
}
