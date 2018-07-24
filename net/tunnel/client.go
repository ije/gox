package tunnel

import (
	"fmt"
	"net"
	"time"
)

type Client struct {
	Server      string
	TunnelName  string
	ForwardPort uint16
}

func (client *Client) Run() {
	for {
		conn, err := client.dialWithHandshake("hello", 15*time.Second)
		if err != nil {
			log.Warn(err)
			continue
		}

		client.heartBeat(conn)
	}
}

func (client *Client) heartBeat(conn net.Conn) {
	for {
		var beatMessage byte
		if dotimeout(func() (err error) {
			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
				return
			}

			beatMessage = buf[0]
			return
		}, 5*time.Second) != nil {
			conn.Close()
			return
		}

		if beatMessage != 1 && beatMessage != 2 {
			conn.Close()
			return
		}

		var retMessage byte = 1
		if beatMessage == 2 {
			if client.dialAndProxy() != nil {
				retMessage = 0
			}
		}

		if dotimeout(func() (err error) {
			_, err = conn.Write([]byte{retMessage})
			return
		}, 5*time.Second) != nil {
			conn.Close()
			return
		}
	}
}

func (client *Client) dialWithHandshake(handshakeMessage string, timeout time.Duration) (conn net.Conn, err error) {
	err = dotimeout(func() (err error) {
		c, err := dial("tcp", client.Server)
		if err != nil {
			return
		}

		err = sendMessage(c, handshakeMessage, []byte(client.TunnelName))
		if err != nil {
			return
		}

		buf := make([]byte, 1)
		_, err = c.Read(buf)
		if err != nil {
			return
		}

		if buf[0] != 1 {
			err = fmt.Errorf("server error")
			return
		}

		conn = c
		return
	}, timeout)
	return
}

func (client *Client) dialAndProxy() (err error) {
	var localConn net.Conn
	var serverConn net.Conn

	err = dotimeout(func() (err error) {
		conn, err := dial("tcp", fmt.Sprintf(":%d", client.ForwardPort))
		if err != nil {
			return
		}

		localConn = conn
		return
	}, time.Second)
	if err != nil {
		log.Warnf("proxy tunnel(%s): dial local failed: %v", client.TunnelName, err)
		return
	}

	serverConn, err = client.dialWithHandshake("proxy", 5*time.Second)
	if err != nil {
		localConn.Close()
		log.Warnf("proxy tunnel(%s): dial server failed: %v", client.TunnelName, err)
		return
	}

	go proxy(serverConn, localConn, 0)
	return
}
