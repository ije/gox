package rpc

import (
	"log"
	"net"
	"net/rpc"
	"regexp"
)

type Server struct {
	*rpc.Server
}

func ListenAndServe(addr string, init func(*Server)) (err error) {
	network := "tcp"
	address := regexp.MustCompile(`^(tcp[46]?|unix(packet)?)@`).ReplaceAllStringFunc(addr, func(net string) string {
		network = net[:len(net)-1]
		return ""
	})
	l, err := net.Listen(network, address)
	if err != nil {
		return
	}
	defer l.Close()

	rpcServer := rpc.NewServer()
	if init != nil {
		init(&Server{rpcServer})
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Println("rpc: Net accept error:", err)
				continue
			}
			return err
		}
		go rpcServer.ServeConn(conn)
	}
}
