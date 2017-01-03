package tunnel

import (
	"io"
	"net"
)

type Client struct {
	Server      string
	AESKey      string
	ServiceName string
	ServicePort uint16
}

func (client *Client) Listen() error {
	for {
		conn, err := dial("tcp", client.Server, client.AESKey)
		if err != nil {
			return err
		}

		err = client.handleConn(conn)
		if err != nil && err != io.EOF {
			return err
		}

		log.Debugf("server connection was colosed, try to reconnect to server(%s)...", client.Server)
	}

	return nil
}

func (client *Client) handleConn(conn net.Conn) (err error) {
	defer conn.Close()

	err = sendData(conn, "hello", []byte(client.ServiceName))
	if err != nil {
		return
	}

	flag, data, err := parseData(conn)
	if err != nil {
		return
	}

	if flag != "hello" {
		err = errf("invalid handshake message")
		return
	}

	if string(data) != client.ServiceName {
		err = errf("invalid handshake message")
		return
	}

	proxyConn, err := dial("tcp", strf(":%d", client.ServicePort), "")
	if err != nil {
		return
	}
	defer proxyConn.Close()

	closeChan := make(chan struct{}, 1)

	go func(proxyConn net.Conn, conn net.Conn, Exit chan struct{}) {
		io.Copy(proxyConn, conn)
		closeChan <- struct{}{}
	}(proxyConn, conn, closeChan)

	go func(proxyConn net.Conn, conn net.Conn, Exit chan struct{}) {
		io.Copy(conn, proxyConn)
		closeChan <- struct{}{}
	}(proxyConn, conn, closeChan)

	<-closeChan
	return
}
