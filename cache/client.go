package cache

import (
	"net/rpc"
	"strconv"
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

func NewClient(addr string, maxConnects int) *Client {
	if maxConnects < 1 {
		maxConnects = 1
	}
	return &Client{addr, make(chan *rpc.Client, maxConnects)}
}

func (client *Client) Service(name string) *Service {
	return &Service{name, client}
}

func (service *Service) Call(method string, args RPCArgs, reply interface{}) (err error) {
	var (
		network   string
		address   string
		rpcClient *rpc.Client
		callTimes int
	)
	select {
	case <-time.After(time.Millisecond):
		network, address = splitNA(service.addr)
		rpcClient, err = rpc.Dial(network, address)
		if err != nil {
			return
		}
	case rpcClient = <-service.rpcClients:
	}
CALL:
	callTimes++
	if err = rpcClient.Call(service.name+"."+method, args, reply); err == rpc.ErrShutdown {
		if callTimes > 3 {
			return
		}
		network, address = splitNA(service.addr)
		rpcClient, err = rpc.Dial(network, address)
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

type RPCArgs map[string]interface{}

func (args RPCArgs) String(key string) (arg string, ok bool) {
	v, ok := args[key]
	if ok {
		arg, ok = v.(string)
	}
	return
}

func (args RPCArgs) Int(key string) (arg int, ok bool) {
	v, ok := args[key]
	if ok {
		switch i := v.(type) {
		case string:
			var err error
			if arg, err = strconv.Atoi(i); err != nil {
				ok = false
			}
		case int:
			arg = i
		case int64:
			arg = int(i)
		case float64:
			arg = int(i)
		default:
			ok = false
		}
	}
	return
}

func (args RPCArgs) Int64(key string) (arg int64, ok bool) {
	v, ok := args[key]
	if ok {
		switch i := v.(type) {
		case string:
			var err error
			if arg, err = strconv.ParseInt(i, 10, 64); err != nil {
				ok = false
			}
		case int:
			arg = int64(i)
		case int64:
			arg = i
		case float64:
			arg = int64(i)
		default:
			ok = false
		}
	}
	return
}

func (args RPCArgs) Bytes(key string) (arg []byte, ok bool) {
	v, ok := args[key]
	if ok {
		arg, ok = v.([]byte)
	}
	return
}

func (args RPCArgs) StringArray(key string) (arg []string, ok bool) {
	v, ok := args[key]
	if ok {
		arg, ok = v.([]string)
	}
	return
}

func (args RPCArgs) IntArray(key string) (arg []int, ok bool) {
	v, ok := args[key]
	if ok {
		arg, ok = v.([]int)
	}
	return
}
