package tunnel

import (
	"net"
	"time"
)

type Client struct {
	Server       string
	ServerSecret string
	TunnelName   string
	ForwardPort  uint16
}

func (client *Client) Run() {
	for {
		conn, err := client.dialWithHandshake("hello", 6*time.Second)
		if err != nil {
			log.Warn(err)
			continue
		}

		client.heartBeat(conn)
	}
}

func (client *Client) heartBeat(conn net.Conn) {
	for {
		var msg byte
		if dotimeout(func() (err error) {
			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
				return
			}

			msg = buf[0]
			return
		}, 3*time.Second) != nil {
			conn.Close()
			return
		}

		if msg != 1 && msg != 2 {
			conn.Close()
			return
		}

		if msg == 2 {
			client.dialAndProxy()
		}
		if dotimeout(func() (err error) {
			_, err = conn.Write([]byte{1})
			return
		}, 3*time.Second) != nil {
			conn.Close()
			return
		}
	}
}

func (client *Client) dialWithHandshake(handshakeMessage string, timeout time.Duration) (conn net.Conn, err error) {
	err = dotimeout(func() (err error) {
		c, err := dial("tcp", client.Server, client.ServerSecret)
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
			err = errf("server error")
			return
		}

		conn = c
		return
	}, timeout)
	return
}

func (client *Client) dialAndProxy() {
	var localConn net.Conn
	if err := dotimeout(func() (err error) {
		conn, err := dial("tcp", strf(":%d", client.ForwardPort), "")
		if err != nil {
			return
		}

		localConn = conn
		return
	}, time.Second); err != nil {
		log.Warnf("proxy tunnel(%s): dial local failed: %v", client.TunnelName, err)
		return
	}

	go func(localConn net.Conn) {
		serverConn, err := client.dialWithHandshake("proxy", 3*time.Second)
		if err != nil {
			log.Warnf("proxy tunnel(%s): dial server failed: %v", client.TunnelName, err)
			localConn.Close()
			return
		}

		proxy(serverConn, localConn)
	}(localConn)
}
