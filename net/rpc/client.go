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

func Dial(addr string, maxConnects int) *Client {
	if maxConnects < 1 {
		maxConnects = 1
	}
	return &Client{addr, make(chan *rpc.Client, maxConnects)}
}

func (client *Client) Call(serviceMethod string, args, reply interface{}) (err error) {
	var rpcClient *rpc.Client
	select {
	case <-time.After(time.Millisecond):
		rpcClient, err = newRPCClient(client.addr)
		if err != nil {
			return
		}
	case rpcClient = <-client.rpcClients:
	}
	var tryTimes int
CALL:
	tryTimes++
	if err = rpcClient.Call(serviceMethod, args, reply); err == rpc.ErrShutdown {
		if tryTimes > 3 {
			return
		}
		rpcClient, err = newRPCClient(client.addr)
		if err != nil {
			return
		}
		goto CALL
	}
	if len(client.rpcClients) < cap(client.rpcClients) {
		client.rpcClients <- rpcClient
	} else {
		rpcClient.Close()
	}
	return
}

func (client *Client) Close() {
	for rpcClient := range client.rpcClients {
		rpcClient.Close()
	}
}

func newRPCClient(addr string) (*rpc.Client, error) {
	network := "tcp"
	addr = regexp.MustCompile(`^(tcp[46]?|unix(packet)?)@`).ReplaceAllStringFunc(addr, func(net string) string {
		network = net[:len(net)-1]
		return ""
	})
	return rpc.Dial(network, addr)
}
