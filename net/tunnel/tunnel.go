package tunnel

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/ije/gox/utils"
)

type TunnelProps struct {
	Name             string
	Port             uint16
	MaxProxyLifetime uint32
}

type Tunnel struct {
	*TunnelProps
	lock       sync.Mutex
	crtime     int64
	online     bool
	clientAddr string
	olTimer    *time.Timer
	connQueue  chan net.Conn
	connPool   chan net.Conn
	listener   net.Listener
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

		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(time.Minute)
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
	t.olTimer = time.AfterFunc(2*time.Duration(heartBeatInterval)*time.Second, func() {
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
	proxyConn(conn1, conn2, time.Duration(t.MaxProxyLifetime)*time.Second)
}

func proxyConn(conn1 net.Conn, conn2 net.Conn, timeout time.Duration) (err error) {
	ec := make(chan error, 1)

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
		case err = <-ec:
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
