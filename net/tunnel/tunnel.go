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

type Tunnel struct {
	Name             string `json:"name"`
	Port             uint16 `json:"port"`
	ProxyConnections int    `json:"proxyConnections"`
	ProxyMaxLifetime int    `json:"proxyMaxLifetime,omitempty"`
	Client           string `json:"client,omitempty"`
	Online           bool   `json:"online"`
	MaxConnections   int    `json:"maxConnections"`
	olTimer          *time.Timer
	connQueue        chan net.Conn
	connPool         chan net.Conn
}

func (t *Tunnel) Serve() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", t.Port))
	if err != nil {
		return
	}

	go listen(l, func(conn net.Conn) {
		log.Debugf("tunnel(%s, Online:%v) new connection ", t.Name, t.Online)

		if !t.Online || len(t.connQueue) >= t.MaxConnections {
			conn.Close()
			return
		}

		t.connQueue <- conn
	})

	return
}

func (t *Tunnel) activate(remoteAddr string, lifetime time.Duration) {
	if t.olTimer != nil {
		t.olTimer.Stop()
	}
	t.Online = true
	t.Client = remoteAddr
	t.olTimer = time.AfterFunc(lifetime, func() {
		t.olTimer = nil
		t.Online = false
		t.Client = ""
	})
}

func (t *Tunnel) proxy(conn1 net.Conn, conn2 net.Conn) {
	t.ProxyConnections++
	proxy(conn1, conn2, time.Duration(t.ProxyMaxLifetime)*time.Second)
	t.ProxyConnections--
}

func sendMessage(conn net.Conn, flag string, data []byte) (err error) {
	flagLen := len(flag)
	if flagLen > 255 {
		err = fmt.Errorf("invalid flag")
		return
	}

	_, err = conn.Write(XTunnelHead)
	if err != nil {
		return
	}

	_, err = conn.Write([]byte{byte(flagLen)})
	if err != nil {
		return
	}
	_, err = conn.Write([]byte(flag))
	if err != nil {
		return
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(data)))
	_, err = conn.Write(buf)
	if err != nil {
		return
	}

	if len(data) > 0 {
		_, err = io.Copy(conn, bytes.NewReader(data))
	}
	return
}

func parseMessage(conn net.Conn) (flag string, data []byte, err error) {
	buf := make([]byte, 8)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	if !bytes.Equal(buf, XTunnelHead) {
		err = fmt.Errorf("invalid x-tunnel head")
		return
	}

	buf = make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	buf = make([]byte, int(buf[0]))
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	flag = string(buf)

	buf = make([]byte, 4)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	if dl := binary.LittleEndian.Uint32(buf); dl > 0 {
		buf := bytes.NewBuffer(nil)
		_, err = io.CopyN(buf, conn, int64(dl))
		if err == nil {
			data = buf.Bytes()
		}
	}
	return
}
