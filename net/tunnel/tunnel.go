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
	Name             string
	Port             uint16
	MaxConnections   int
	MaxProxyLifetime int
	online           bool
	client           string
	proxyConnections int
	olTimer          *time.Timer
	connQueue        chan net.Conn
	connPool         chan net.Conn
	listener         net.Listener
}

func (t *Tunnel) Serve() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", t.Port))
	if err != nil {
		return
	}

	t.listener = l
	go listen(t.listener, func(conn net.Conn) {
		log.Debugf("tunnel(%s, online:%v) new connection ", t.Name, t.online)

		if !t.online || len(t.connQueue) >= t.MaxConnections {
			conn.Close()
			return
		}

		t.connQueue <- conn
	})
	return
}

func (t *Tunnel) activate(client string) {
	t.online = true
	t.client = client
	if t.olTimer != nil {
		t.olTimer.Stop()
	}
	t.olTimer = time.AfterFunc(15*time.Second, func() {
		t.olTimer = nil
		t.online = false
		t.client = ""
	})
}

func (t *Tunnel) proxy(conn1 net.Conn, conn2 net.Conn) {
	t.proxyConnections++
	proxy(conn1, conn2, time.Duration(t.MaxProxyLifetime)*time.Second)
	t.proxyConnections--
}

func sendMessage(conn net.Conn, flag string, data []byte) (err error) {
	flagLen := len(flag)
	dataLen := uint32(len(data))
	if flagLen > 255 {
		err = fmt.Errorf("invalid flag")
		return
	}

	buf := bytes.NewBuffer(XTunnelHead)
	buf.WriteByte(byte(flagLen))

	p := make([]byte, 4)
	binary.LittleEndian.PutUint32(p, dataLen)
	buf.Write(p)

	buf.WriteString(flag)

	_, err = io.Copy(conn, buf)
	if err != nil {
		return
	}

	if dataLen > 0 {
		_, err = io.Copy(conn, bytes.NewReader(data))
	}
	return
}

func parseMessage(conn net.Conn) (flag string, data []byte, err error) {
	buf := make([]byte, len(XTunnelHead))
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	if !bytes.Equal(buf, XTunnelHead) {
		err = fmt.Errorf("invalid head")
		return
	}

	buf = make([]byte, 5)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}

	fl := int(buf[0])
	dl := binary.LittleEndian.Uint32(buf[1:])

	buf = make([]byte, fl)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}

	flag = string(buf)

	if dl > 0 {
		buf := bytes.NewBuffer(nil)
		_, err = io.CopyN(buf, conn, int64(dl))
		if err == nil {
			data = buf.Bytes()
		}
	}
	return
}
