package sockets

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var cert *tls.Certificate

func loadCertificate() {
	curDir, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Printf("can`t get working directory: %s", err.Error()))
	}
	ca, err := tls.LoadX509KeyPair(curDir+"/certs/client.pem", curDir+"/certs/client.key")
	if err != nil {
		log.Fatal(fmt.Printf("can`t read client certificate: %s", err.Error()))
	}
	cert = &ca
	log.Println("CLIENT CERTIFICATE loaded")
}

func getHttpsConnection(ip string) (conn net.Conn, err error) {
	if cert == nil {
		loadCertificate()
	}

	d := tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout:   3 * time.Second,
			Deadline:  time.Now().Add(time.Second * 5),
			KeepAlive: time.Second * 30,
		},
		Config: &tls.Config{
			Certificates:       []tls.Certificate{*cert},
			InsecureSkipVerify: true,
		},
	}
	conn, err = d.Dial("tcp", ip+":443")
	if err != nil {
		return nil, err
	}

	bytes := bytesPool.Get().([]byte)
	_, err = conn.Read(bytes)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	bytes = bytes[:0]
	bytesPool.Put(bytes)

	return conn, nil
}
