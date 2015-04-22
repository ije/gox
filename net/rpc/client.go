package rpc

import (
	"net/rpc"
	"regexp"
	"time"
)

type Client struct {
	addr       string
	rpcClients chan *rpc.Client
}

type Service struct {
	name string
	*Client
}

func Dial(addr string, maxConnects int) *Client {
	if maxConnects < 1 {
		maxConnects = 1
	}
	return &Client{addr, make(chan *rpc.Client, maxConnects)}
}

func (client *Client) Service(name string) *Service {
	return &Service{name, client}
}

func (service *Service) Call(method string, argument, reply interface{}) (err error) {
	var rpcClient *rpc.Client
	select {
	case <-time.After(time.Millisecond):
		rpcClient, err = newRPCClient(service.addr)
		if err != nil {
			return
		}
	case rpcClient = <-service.rpcClients:
	}
	var callTimes int
CALL:
	callTimes++
	if err = rpcClient.Call(service.name+"."+method, argument, reply); err == rpc.ErrShutdown {
		if callTimes > 3 {
			return
		}
		rpcClient, err = newRPCClient(service.addr)
		if err != nil {
			return
		}
		goto CALL
	}
	if len(service.rpcClients) < cap(service.rpcClients) {
		service.rpcClients <- rpcClient
	}
	return
}

func newRPCClient(addr string) (*rpc.Client, error) {
	network := "tcp"
	addr = regexp.MustCompile(`^(tcp[46]?|unix(packet)?)@`).ReplaceAllStringFunc(addr, func(net string) string {
		network = net[:len(net)-1]
		return ""
	})
	return rpc.Dial(network, addr)
}
