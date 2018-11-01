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
			log.Warnf("tunnel(%s): dial server failed: %v", client.Tunnel.Name, err)
			time.Sleep(time.Second)
			continue
		}

		client.Serve(conn)
	}
}

func (client *Client) Serve(conn net.Conn) {
	for {
		if err := dotimeout(func() (err error) {
			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
				return
			}

			beatMessage := buf[0]
			if beatMessage != 1 && beatMessage != 2 {
				err = fmt.Errorf("server sent a bad data")
				return
			}

			var retMessage byte = 1
			if beatMessage == 2 && client.dialAndProxy() != nil {
				retMessage = 0
			}
			_, err = conn.Write([]byte{retMessage})
			return
		}, 15*time.Second); err != nil {
			log.Warnf("tunnel(%s): %v", client.Tunnel.Name, err)
			conn.Close()
			return
		}
	}
}

func (client *Client) dialAndProxy() (err error) {
	err = dotimeout(func() (err error) {
		localConn, err := dial("tcp", fmt.Sprintf(":%d", client.ForwardPort))
		if err != nil {
			err = fmt.Errorf("dial local failed: %v", err)
			return
		}

		serverConn, err := client.dialWithHandshake("proxy")
		if err != nil {
			err = fmt.Errorf("dial server failed: %v", err)
			return
		}

		go proxy(serverConn, localConn, time.Duration(client.Tunnel.MaxProxyLifetime)*time.Second)
		return
	}, 10*time.Second)
	if err != nil {
		log.Warnf("tunnel(%s): %v", client.Tunnel.Name, err)
	}
	return
}

func (client *Client) dialWithHandshake(flag string) (conn net.Conn, err error) {
	c, err := dial("tcp", client.Server)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(nil)
	if flag == "hello" {
		gob.NewEncoder(buffer).Encode(client.Tunnel)
	} else if flag == "proxy" {
		buffer.WriteString(client.Tunnel.Name)
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
		err = fmt.Errorf("server sent a bad data")
		c.Close()
		return
	}

	conn = c
	return
}
