package tunnel

import (
	"net"
	"time"
)

type Client struct {
	Server      string
	Password    string
	Tunnel      string
	ForwardPort uint16
}

func (client *Client) Run() {
	for {
		client.dial(false)
	}
}

func (client *Client) dial(proxy bool) {
	var conn net.Conn
	var err error

	ec := make(chan error, 1)

	go func() {
		conn, err = dial("tcp", client.Server, client.Password)
		if err != nil {
			log.Warnf("tunnel(%s): dial remote: %v", client.Tunnel, err)
			ec <- err
			return
		}

		msg := "hello"
		if proxy {
			msg = "proxy"
		}
		err = sendMessage(conn, msg, []byte(client.Tunnel))
		if err != nil {
			ec <- err
			return
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err != nil || buf[0] != 1 {
			ec <- err
			return
		}

		ec <- nil
	}()

	select {
	case err := <-ec:
		if err != nil {
			if conn != nil {
				conn.Close()
			}
			return
		}
	case <-time.After(3 * time.Second):
		close(ec)
		if conn != nil {
			conn.Close()
		}
		return
	}

	if proxy {
		client.proxy(conn)
		return
	}

	// heartBeat
	for {
		ec := make(chan error, 1)
		msg := make(chan byte, 1)

		go func() {
			_, err := conn.Write([]byte{'!'})
			if err != nil {
				close(msg)
				ec <- err
				return
			}

			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
				close(msg)
				ec <- err
				return
			}

			ec <- nil
			msg <- buf[0]
		}()

		select {
		case err := <-ec:
			if err != nil {
				conn.Close()
				return
			}
		case <-time.After(3 * time.Second):
			close(ec)
			close(msg)
			conn.Close()
			return
		}

		if <-msg == 1 {
			go client.dial(true)
			return
		}
	}
}

func (client *Client) proxy(conn net.Conn) {
	proxyConn, err := dial("tcp", strf(":%d", client.ForwardPort), "")
	if err != nil {
		log.Warnf("tunnel(%s): dial local failed: %v", client.Tunnel, err)
		conn.Close()
		return
	}

	proxy(conn, proxyConn)
}
