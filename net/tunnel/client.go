package tunnel

import (
	"net"
	"time"
)

type Client struct {
	Server      string
	AESKey      string
	ServiceName string
	ServicePort uint16
	Connections int
}

func (client *Client) Run() {
	for i := 0; i < client.Connections-1; i++ {
		go client.dial()
	}
	client.dial()
}

func (client *Client) dial() {
	for {
		conn, err := dial("tcp", client.Server, client.AESKey)
		if err != nil {
			log.Warnf("service(%s): dial remote: %v", client.ServiceName, err)
			time.Sleep(time.Second)
			continue
		}

		err = sendMessage(conn, "hello", []byte(client.ServiceName))
		if err != nil {
			conn.Close()
			continue
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err != nil || buf[0] != 1 {
			conn.Close()
			continue
		}

		client.heartBeat(conn)
	}
}

func (client *Client) heartBeat(conn net.Conn) {
	for {
		ec := make(chan error, 1)
		msg := make(chan byte, 1)

		go func() {
			_, err := conn.Write([]byte{'!'})
			if err != nil {
				ec <- err
				return
			}

			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
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
			conn.Close()
			return
		}

		if <-msg == 1 {
			go client.proxy(conn)
			return
		}
	}
}

func (client *Client) proxy(conn net.Conn) {
	proxyConn, err := dial("tcp", strf(":%d", client.ServicePort), "")
	if err != nil {
		log.Warnf("service(%s): dial local failed: %v", client.ServiceName, err)
		conn.Close()
		return
	}

	proxy(conn, proxyConn)
}
