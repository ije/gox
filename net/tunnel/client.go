package tunnel

import (
	"io"
	"net"
	"time"
)

type Client struct {
	Server      string
	AESKey      string
	ServiceName string
	ServicePort uint16
	Connections int
}

func (client *Client) Listen() error {
	conn, err := dial("tcp", client.Server, client.AESKey)
	if err != nil {
		return err
	}
	conn.Close()

	for i := 0; i < client.Connections-1; i++ {
		go client.dial()
	}

	client.dial()
	return nil
}

func (client *Client) dial() {
	for {
		conn, err := dial("tcp", client.Server, client.AESKey)
		if err != nil {
			log.Warn("client: dial:", err)
			time.Sleep(time.Second)
			continue
		}

		err = client.handleConn(conn)
		if err != nil && err != io.EOF {
			log.Warn("client: handle connection:", err)
		}
	}
}

func (client *Client) handleConn(conn net.Conn) (err error) {
	err = sendMessage(conn, "hello", []byte(client.ServiceName))
	if err != nil {
		conn.Close()
		return
	}

	ec := make(chan error, 1)
	firstByteChan := make(chan byte, 1)

	go func() {
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			ec <- err
			return
		}

		firstByteChan <- buf[0]
		ec <- nil
	}()

	// connection will be closed when not be used in 30 minutes
	select {
	case err = <-ec:
		if err != nil {
			conn.Close()
			return
		}
	case <-time.After(30 * time.Minute):
		conn.Close()
		return
	}

	go client.proxy(<-firstByteChan, conn)
	return
}

func (client *Client) proxy(firstByte byte, conn net.Conn) {
	proxyConn, err := dial("tcp", strf(":%d", client.ServicePort), "")
	if err != nil {
		log.Warnf("client: service(%s) dail failed: %v", client.ServiceName, err)
		conn.Close()
		return
	}

	_, err = proxyConn.Write([]byte{firstByte})
	if err != nil {
		log.Warnf("client: service(%s) write first byte failed: %v", client.ServiceName, err)
		conn.Close()
		return
	}

	proxy(conn, proxyConn)
}
