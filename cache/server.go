package cache

import (
	"encoding/gob"
	"log"
	"net"
	"net/rpc"
	"regexp"
)

func ListenAndServe(addr string, init func(*rpc.Server), gobTypes ...interface{}) (err error) {
	network, address := splitNA(addr)
	l, err := net.Listen(network, address)
	if err != nil {
		return
	}
	defer l.Close()

	for _, val := range gobTypes {
		switch v := val.(type) {
		case map[string]interface{}:
			if len(v) > 0 {
				for name, value := range v {
					gob.RegisterName(name, value)
				}
				break
			}
			gob.Register(v)
		default:
			gob.Register(v)
		}
	}

	rpcSrv := rpc.NewServer()
	if init != nil {
		rpcSrv.RegisterName("Cache", &cacheRPCServer{})
		init(rpcSrv)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Println("cache: Net accept error:", err)
				continue
			}
			return err
		}
		go rpcSrv.ServeConn(conn)
	}
}

func splitNA(addr string) (network, address string) {
	network = "tcp"
	address = regexp.MustCompile(`^(tcp[46]?|unix(packet)?)@`).ReplaceAllStringFunc(addr, func(net string) string {
		network = net[:len(net)-1]
		return ""
	})
	return
}
