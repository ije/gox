package tunnel

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
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

		// fmt.Println("heartbeat returns:", flag)
		if flag == FlagProxy {
			go client.dialAndProxy()
		} else if flag != FlagHello {
			return
		}
	}
}

func (client *Client) dialAndProxy() {
	localConn, err := net.Dial("tcp", fmt.Sprintf(":%d", client.ForwardPort))
	if err != nil {
		return
	}

	serverConn, err := client.dial(FlagProxy)
	if err != nil {
		log.Println("[dialAndProxy] can not dial server: ", err)
		localConn.Close()
		return
	}

	go proxy(serverConn, localConn, time.Duration(client.Tunnel.MaxProxyLifetime)*time.Second)
	return
}

func (client *Client) dial(flag Flag) (conn net.Conn, err error) {
	c, err := net.Dial("tcp", client.Server)
	if err != nil {
		return
	}

	buffer := bytes.NewBuffer(nil)
	if flag == FlagHello {
		gob.NewEncoder(buffer).Encode(TunnelInfo{
			Name:             client.Tunnel.Name,
			Port:             client.Tunnel.Port,
			MaxProxyLifetime: client.Tunnel.MaxProxyLifetime,
		})
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
