package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"gox/utils"
)

type Client struct {
	Server      string
	Tunnel      *TunnelProps
	ForwardPort uint16
}

func (client *Client) Connect() {
	for {
		conn, err := client.dial(FlagHello)
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
		flag, _, err := parseMessage(conn)
		if err != nil {
			return
		}

		log.Println("heartbeat flag:", flag)
		if flag == FlagHello {
			err = sendMessage(conn, FlagHello, nil)
		} else if flag == FlagProxy {
			err2 := client.dialAndProxy()
			if err2 != nil {
				err = sendMessage(conn, FlagError, []byte(err2.Error()))
			} else {
				err = sendMessage(conn, FlagReady, nil)
			}
		} else {
			return
		}
		if err != nil {
			return
		}
	}
}

func (client *Client) dialAndProxy() (err error) {
	localConn, err := net.Dial("tcp", fmt.Sprintf(":%d", client.ForwardPort))
	if err != nil {
		err = fmt.Errorf("dial local: %v", err)
		return
	}

	serverConn, err := client.dial(FlagProxy)
	if err != nil {
		localConn.Close()
		err = fmt.Errorf("dial server: %v", err)
		return
	}

	go utils.Proxy(serverConn, localConn, time.Duration(client.Tunnel.MaxProxyLifetime)*time.Second)
	return
}

func (client *Client) dial(flag Flag) (conn net.Conn, err error) {
	c, err := net.Dial("tcp", client.Server)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(nil)
	if flag == FlagHello {
		buffer.WriteByte(byte(len(client.Tunnel.Name)))
		buffer.WriteString(client.Tunnel.Name)
		p := make([]byte, 2)
		binary.LittleEndian.PutUint16(p, client.Tunnel.Port)
		buffer.Write(p)
		p = make([]byte, 4)
		binary.LittleEndian.PutUint32(p, client.Tunnel.MaxProxyLifetime)
		buffer.Write(p)
	} else if flag == FlagProxy {
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

	conn = c
	return
}
