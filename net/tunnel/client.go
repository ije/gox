package tunnel

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"time"
)

type Client struct {
	Server      string
	Tunnel      Tunnel
	ForwardPort uint16
}

func (client *Client) Connect() {
	for {
		conn, err := client.dialWithHandshake("hello")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		client.serveHeartBeat(conn)
	}
}

func (client *Client) serveHeartBeat(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			return
		}

		if buf[0] == 2 {
			client.dialAndProxy()
		}
	}
}

func (client *Client) dialAndProxy() (err error) {
	localConn, err := net.Dial("tcp", fmt.Sprintf(":%d", client.ForwardPort))
	if err != nil {
		return
	}

	serverConn, err := client.dialWithHandshake("proxy")
	if err != nil {
		err = fmt.Errorf("dial server for proxy request: %v", err)
		localConn.Close()
		return
	}

	go proxy(serverConn, localConn, time.Duration(client.Tunnel.MaxProxyLifetime)*time.Second)
	return
}

func (client *Client) dialWithHandshake(flag string) (conn net.Conn, err error) {
	c, err := net.Dial("tcp", client.Server)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(nil)
	if flag == "hello" {
		gob.NewEncoder(buffer).Encode(client.Tunnel)
	} else if flag == "proxy" {
		buffer.WriteString(client.Tunnel.Name)
	} else {
		err = fmt.Errorf("invalid flag")
		c.Close()
		return
	}

	err = sendMessage(c, flag, buffer.Bytes())
	if err != nil {
		c.Close()
		return
	}

	buf := make([]byte, 1)
	_, err = c.Read(buf)
	if err != nil {
		c.Close()
		return
	}

	if buf[0] != 1 {
		err = fmt.Errorf("server error")
		c.Close()
		return
	}

	conn = c
	return
}
