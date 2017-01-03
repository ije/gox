package tunnel

import (
	"encoding/binary"
	"io"
	"net"
)

type Client struct {
	Name   string
	Server string
	AESKey string
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

	err = sendData(conn, "hello", []byte(client.Name))
	if err != nil {
		return
	}

	flag, data, err := parseData(conn)
	if err != nil {
		return
	}

	if flag != "hello" {
		err = errf("invalid handshake message: %s", flag)
		return
	}

	if len(data) != 2 {
		err = errf("invalid proxy port data")
		return
	}

	proxyPort := binary.LittleEndian.Uint16(data)
	proxyConn, err := dial("tcp", strf(":%d", proxyPort), "")
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
