package tunnel

import (
	"io"
	"net"
	"time"
)

type Client struct {
	Server      string
	AESKey      string
	ServiceName string
	ServicePort uint16
}

func (client *Client) Listen() error {
	conn, err := dial("tcp", client.Server, client.AESKey)
	if err != nil {
		return err
	}
	conn.Close()

	for {
		conn, err := dial("tcp", client.Server, client.AESKey)
		if err != nil {
			log.Warnf("x.tunnel.client: dial:", err)
			time.Sleep(time.Second)
			continue
		}

		err = client.handleConn(conn)
		if err != nil && err != io.EOF {
			log.Warnf("x.tunnel.client: handle connection:", err)
			continue
		}
	}

	return nil
}

func (client *Client) handleConn(conn net.Conn) (err error) {
	err = sendData(conn, "hello", []byte(client.ServiceName))
	if err != nil {
		conn.Close()
		return
	}

	ec := make(chan error, 1)

	go func() {
		flag, data, err := parseData(conn)
		if err != nil {
			ec <- err
			return
		}

		if flag != "start-proxy" || string(data) != client.ServiceName {
			ec <- errf("invalid handshake message")
			return
		}

		ec <- nil
	}()

	select {
	case err = <-ec:
		if err != nil {
			conn.Close()
			return
		}
	case <-time.After(time.Minute):
		conn.Close()
		return
	}

	go client.proxy(conn)
	return
}

func (client *Client) proxy(conn net.Conn) {
	proxyConn, err := dial("tcp", strf(":%d", client.ServicePort), "")
	if err != nil {
		log.Warnf("x.tunnel.client: dail service(%s) failed: %v", client.ServiceName, err)
		conn.Close()
		return
	}

	proxy(conn, proxyConn)
}
