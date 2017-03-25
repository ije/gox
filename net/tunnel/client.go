package tunnel

import (
	"net"
	"time"

	"github.com/ije/gox/utils"
)

type Client struct {
	Server      string
	AESKey      string
	TunnelName  string
	TunnelPort  uint16
	LocalPort   uint16
	Connections int
	registered  bool
}

func (client *Client) Run() (err error) {
	err = client.register()
	if err != nil {
		return
	}

	for i := 0; i < client.Connections-1; i++ {
		go client.link()
	}
	client.link()
	return
}

func (client *Client) register() (err error) {
	if client.registered {
		return
	}

	conn, err := dial("tcp", client.Server, client.AESKey)
	if err != nil {
		log.Warnf("tunnel(%s): dial remote: %v", client.TunnelName, err)
		return
	}
	defer conn.Close()

	err = sendMessage(conn, "register", []byte(utils.MustEncodeGobBytes(client)))
	if err != nil {
		log.Warnf("tunnel(%s): %v", client.TunnelName, err)
		return
	}

	flag, data, err := parseMessage(conn)
	if err != nil {
		log.Warnf("tunnel(%s): %v", client.TunnelName, err)
		return
	}

	if flag == "error" {
		err = errf(string(data))
	} else if flag != "registered" {
		err = errf("invalid flag")
	}

	if err != nil {
		log.Warnf("tunnel(%s): register tunnel: %v", client.TunnelName, err)
		return
	}

	client.registered = true
	return
}

func (client *Client) link() {
	for {
		conn, err := dial("tcp", client.Server, client.AESKey)
		if err != nil {
			log.Warnf("tunnel(%s): dial remote: %v", client.TunnelName, err)
			time.Sleep(time.Second)
			continue
		}

		err = sendMessage(conn, "hello", []byte(client.TunnelName))
		if err != nil {
			conn.Close()
			continue
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err != nil || buf[0] != 1 {
			conn.Close()
			continue
		}

		client.heartBeat(conn)
	}
}

func (client *Client) heartBeat(conn net.Conn) {
	for {
		ec := make(chan error, 1)
		ok := make(chan byte, 1)

		go func() {
			_, err := conn.Write([]byte{'!'})
			if err != nil {
				ec <- err
				return
			}

			buf := make([]byte, 1)
			_, err = conn.Read(buf)
			if err != nil {
				ec <- err
				return
			}

			ec <- nil
			ok <- buf[0]
		}()

		select {
		case err := <-ec:
			if err != nil {
				conn.Close()
				return
			}
		case <-time.After(3 * time.Second):
			// heartbeat timeout
			conn.Close()
			return
		}

		// start forward
		if <-ok == 1 {
			go client.forward(conn)
			return
		}
	}
}

func (client *Client) forward(conn net.Conn) {
	localConn, err := dial("tcp", strf(":%d", client.LocalPort), "")
	if err != nil {
		log.Warnf("tunnel(%s): dial local: %v", client.TunnelName, err)
		conn.Close()
		return
	}

	proxy(conn, localConn)
}
