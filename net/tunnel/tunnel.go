package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/ije/gox/utils"
)

var XTunnelHead = []byte("X-TUNNEL")

type Tunnel struct {
	Name             string
	Port             uint16
	MaxProxyLifetime int
	lock             sync.Mutex
	crtime           int64
	online           bool
	clientAddr       string
	proxyConnections int
	olTimer          *time.Timer
	connQueue        chan net.Conn
	connPool         chan net.Conn
	listener         net.Listener
}

func (t *Tunnel) ListenAndServe() (err error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.Port))
	if err != nil {
		return
	}
	defer listener.Close()

	t.listener = listener
	for {
		conn, err := listener.Accept()
		if err != nil {
			t.listener = nil
			return err
		}

		go t.handleConn(conn)
	}
}

func (t *Tunnel) handleConn(conn net.Conn) {
	if !t.online {
		conn.Close()
		return
	}

	t.connQueue <- conn
}

func (t *Tunnel) activate(addr net.Addr) {
	t.lock.Lock()
	defer t.lock.Unlock()

	remoteAddr, _ := utils.SplitByLastByte(addr.String(), ':')
	t.online = true
	t.clientAddr = remoteAddr
	if t.olTimer != nil {
		t.olTimer.Stop()
		t.olTimer = nil
	}
	t.olTimer = time.AfterFunc(2*heartBeatInterval*time.Second, func() {
		t.olTimer = nil
		t.online = false
		t.clientAddr = ""
	})
}

func (t *Tunnel) unactivate() {
	t.lock.Lock()
	defer t.lock.Unlock()

	if t.olTimer != nil {
		t.olTimer.Stop()
		t.olTimer = nil
	}
	t.online = false
	t.clientAddr = ""
}

func (t *Tunnel) close() {
	t.unactivate()
	close(t.connQueue)
	close(t.connPool)
	if l := t.listener; l != nil {
		t.listener = nil
		l.Close()
	}
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
