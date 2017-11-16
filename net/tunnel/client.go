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
		conn, err := client.dialWithHandshake("hello", 6*time.Second)
		if err != nil {
			log.Warn(err)
			continue
		}

		client.heartBeat(conn)
	}
}

func (client *Client) heartBeat(conn net.Conn) {
	for {
		mc := make(chan byte, 1)
		ec := make(chan error, 1)

		go func(conn net.Conn, mc chan byte, ec chan error) {
			buf := make([]byte, 1)
			_, err := conn.Read(buf)
			if err != nil {
				ec <- err
				return
			}

			mc <- buf[0]
			ec <- nil
		}(conn, mc, ec)

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

		msg := <-mc
		if msg == 2 {
			client.dialAndProxy()
		} else if msg != 1 {
			conn.Close()
			return
		}

		ec = make(chan error, 1)
		go func(conn net.Conn, ec chan error) {
			_, err := conn.Write([]byte{1})
			if err != nil {
				ec <- err
				return
			}

			ec <- nil
		}(conn, ec)

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
	}
}

func (client *Client) dialWithHandshake(handshakeMessage string, timeout time.Duration) (conn net.Conn, err error) {
	cc := make(chan net.Conn, 1)
	ec := make(chan error, 1)

	go func(cc chan net.Conn, ec chan error, msg string) {
		c, e := dial("tcp", client.Server, client.ServerSecret)
		if e != nil {
			ec <- e
			return
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
	}(cc, ec, handshakeMessage)

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

func (client *Client) dialAndProxy() {
	ec := make(chan error, 1)
	cc := make(chan net.Conn, 1)
	go func(cc chan net.Conn, ec chan error) {
		conn, err := dial("tcp", strf(":%d", client.ForwardPort), "")
		if err != nil {
			ec <- err
			return
		}

		cc <- conn
		ec <- nil
	}(cc, ec)

	var localConn net.Conn
	select {
	case err := <-ec:
		if err != nil {
			log.Warnf("proxy tunnel(%s): dial local failed: %v", client.TunnelName, err)
			return
		}

		localConn = <-cc
	case <-time.After(time.Second):
		log.Warnf("proxy tunnel(%s): dial local timeout", client.TunnelName)
		return
	}

	go func(localConn net.Conn) {
		serverConn, err := client.dialWithHandshake("proxy", 3*time.Second)
		if err != nil {
			log.Warnf("proxy tunnel(%s): dial server failed: %v", client.TunnelName, err)
			localConn.Close()
			return
		}

		proxy(serverConn, localConn)
	}(localConn)
}
