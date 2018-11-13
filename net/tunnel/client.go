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
			log.Warnf("tunnel(%s): connect: %v", client.Tunnel.Name, err)
			time.Sleep(time.Second)
			continue
		}

		client.heartBeat(conn)
	}
}

func (client *Client) heartBeat(conn net.Conn) {
	for {
		if err := dotimeout(func() (err error) {
			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
				return
			}

			beatState := buf[0]
			if beatState != 1 && beatState != 2 {
				err = fmt.Errorf("server error")
				return
			}

			var retState byte = 1
			if beatState == 2 && client.dialAndProxy() != nil {
				retState = 0
			}
			_, err = conn.Write([]byte{retState})
			return
		}, 15*time.Second); err != nil {
			log.Warnf("tunnel(%s) heart beat: %v", client.Tunnel.Name, err)
			conn.Close()
			return
		}
	}
}

func (client *Client) dialAndProxy() (err error) {
	if err = dotimeout(func() (err error) {
		var localConn net.Conn
		for i := 0; i < 6; i++ {
			localConn, err = dial("tcp", fmt.Sprintf(":%d", client.ForwardPort))
			if err == nil {
				break
			}
			if i < 5 {
				time.Sleep(time.Second / 2)
			}
		}
		if err != nil {
			err = fmt.Errorf("dial local failed: %v", err)
			return
		}

		serverConn, err := client.dialWithHandshake("proxy")
		if err != nil {
			localConn.Close()
			return
		}

		go proxy(serverConn, localConn, time.Duration(client.Tunnel.MaxProxyLifetime)*time.Second)
		return
	}, 15*time.Second); err != nil {
		log.Warnf("tunnel(%s) dialAndProxy: %v", client.Tunnel.Name, err)
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
	}

	conn = c
	return
}
