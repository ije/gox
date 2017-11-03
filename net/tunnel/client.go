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
	Connections int
	connQueue   chan struct{}
}

func (client *Client) Run() {
	connections := client.Connections
	if connections < 1 {
		connections = 1
	}
	client.connQueue = make(chan struct{}, connections)

	for {
		client.connQueue <- struct{}{}
		go client.dial()
	}
}

func (client *Client) dial() {
	conn, err := dial("tcp", client.Server, client.Password)
	if err != nil {
		log.Warnf("tunnel(%s): dial remote: %v", client.Tunnel, err)
		<-client.connQueue
		return
	}

	ec := make(chan error, 1)

	go func() {
		err = sendMessage(conn, "hello", []byte(client.Tunnel))
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
			<-client.connQueue
			conn.Close()
			return
		}
	case <-time.After(3 * time.Second):
		close(ec)
		<-client.connQueue
		conn.Close()
		return
	}

	go client.heartBeat(conn)
}

func (client *Client) heartBeat(conn net.Conn) {
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
				<-client.connQueue
				conn.Close()
				return
			}
		case <-time.After(3 * time.Second):
			close(ec)
			close(msg)
			<-client.connQueue
			conn.Close()
			return
		}

		if <-msg == 1 {
			<-client.connQueue
			go client.proxy(conn)
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
