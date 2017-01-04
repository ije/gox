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
			time.Sleep(100 * time.Millisecond)
			continue
		}

		err = client.handleConn(conn)
		if err != nil && err != io.EOF {
			log.Warnf("x.tunnel.client: handle conn:", err)
			continue
		}
	}

	return nil
}

func (client *Client) handleConn(conn net.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	err = sendData(conn, "hello", []byte(client.ServiceName))
	if err != nil {
		return
	}

	flag, data, err := parseData(conn)
	if err != nil {
		return
	}

	if flag != "start proxy" || string(data) != client.ServiceName {
		err = errf("invalid handshake message")
		return
	}

	proxyConn, err := dial("tcp", strf(":%d", client.ServicePort), "")
	if err != nil {
		return
	}

	go proxy(conn, proxyConn)
	return
}
