package tunnel

import (
	"net"
	"time"
)

type Client struct {
	Server       string
	ServerSecret string
	TunnelName   string
	ForwardPort  uint16
}

func (client *Client) Run() {
	for {
		client.dial(false)
	}
}

func (client *Client) dial(proxy bool) {
	conn, err := client.handshake(proxy, 6*time.Second)
	if err != nil {
		log.Warn("handshake:", err)
		return
	}

	if proxy {
		client.proxy(conn)
		return
	}

	client.heartBeat(conn)
}

func (client *Client) heartBeat(conn net.Conn) {
	for {
		mc := make(chan byte, 1)
		ec := make(chan error, 1)

		go func(mc chan byte, ec chan error) {
			buf := make([]byte, 1)
			_, err := conn.Read(buf)
			if err != nil {
				ec <- err
				return
			}

			mc <- buf[0]
			ec <- nil
		}(mc, ec)

		select {
		case err := <-ec:
			if err != nil {
				conn.Close()
				return
			}
		case <-time.After(3 * time.Second):
			conn.Close()
			return
		}

		if msg := <-mc; msg == 2 {
			go client.dial(true)
		} else if msg != 1 {
			conn.Close()
			return
		}
	}
}

func (client *Client) handshake(proxy bool, timeout time.Duration) (conn net.Conn, err error) {
	cc := make(chan net.Conn, 1)
	ec := make(chan error, 1)

	go func(cc chan net.Conn, ec chan error) {
		c, e := dial("tcp", client.Server, client.ServerSecret)
		if e != nil {
			ec <- e
			return
		}

		msg := "hello"
		if proxy {
			msg = "proxy"
		}
		e = sendMessage(c, msg, []byte(client.TunnelName))
		if e != nil {
			ec <- e
			return
		}

		buf := make([]byte, 1)
		_, e = c.Read(buf)
		if e != nil || buf[0] != 1 {
			ec <- e
			return
		}

		cc <- c
		ec <- nil
	}(cc, ec)

	select {
	case err = <-ec:
		if err == nil {
			conn = <-cc
		}
	case <-time.After(timeout):
		err = errf("time out")
	}

	return
}

func (client *Client) proxy(conn net.Conn) {
	proxyConn, err := dial("tcp", strf(":%d", client.ForwardPort), "")
	if err != nil {
		log.Warnf("tunnel(%s): dial local failed: %v", client.TunnelName, err)
		conn.Close()
		return
	}

	proxy(conn, proxyConn)
}
