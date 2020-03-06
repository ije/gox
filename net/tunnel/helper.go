package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

var XTunnelHead = []byte("X-TUNNEL")

func sendMessage(conn net.Conn, flag Flag, data []byte) (err error) {
	buf := bytes.NewBuffer(XTunnelHead)
	buf.WriteByte(byte(flag))

	p := make([]byte, 4)
	dl := uint32(len(data))
	binary.LittleEndian.PutUint32(p, dl)
	buf.Write(p)
	if dl > 0 {
		buf.Write(data)
	}

	_, err = io.Copy(conn, buf)
	return
}

func parseMessage(conn net.Conn) (flag Flag, data []byte, err error) {
	// check head
	buf := make([]byte, len(XTunnelHead))
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	if !bytes.Equal(buf, XTunnelHead) {
		err = fmt.Errorf("invalid head")
		return
	}

	// parse flag
	buf = make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	flag = Flag(buf[0])

	// parse data
	buf = make([]byte, 4)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	dl := binary.LittleEndian.Uint32(buf)
	if dl > 0 {
		buf := bytes.NewBuffer(nil)
		_, err = io.CopyN(buf, conn, int64(dl))
		if err == nil {
			data = buf.Bytes()
		}
	}
	return
}

func proxy(conn1 net.Conn, conn2 net.Conn, timeout time.Duration) (err error) {
	ec := make(chan error, 2)

	go func(conn1 net.Conn, conn2 net.Conn, ec chan error) {
		_, err := io.Copy(conn1, conn2)
		ec <- err
	}(conn1, conn2, ec)

	go func(conn1 net.Conn, conn2 net.Conn, ec chan error) {
		_, err := io.Copy(conn2, conn1)
		ec <- err
	}(conn1, conn2, ec)

	if timeout > 0 {
		select {
		case e := <-ec:
			err = e
		case <-time.After(timeout):
			err = fmt.Errorf("timeout")
		}
	} else {
		err = <-ec
	}

	conn1.Close()
	conn2.Close()
	return
}

type TunnelInfo struct {
	Name             string `json:"name"`
	Port             uint16 `json:"port"`
	MaxProxyLifetime int    `json:"maxProxyLifetime,omitempty"`
	Online           bool   `json:"online"`
	ClientAddr       string `json:"clientAddr"`
	ProxyConnections int    `json:"proxyConnections"`
}

type TunnelSlice []TunnelInfo

func (p TunnelSlice) Len() int           { return len(p) }
func (p TunnelSlice) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p TunnelSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
