package tunnel

import (
	"fmt"
	"net"
	"sync"
	"time"

	"gox/utils"
)

type TunnelProps struct {
	Name             string
	Port             uint16
	MaxProxyLifetime uint32
}

type Tunnel struct {
	*TunnelProps
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

		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			tcpConn.SetKeepAlive(false)
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
	t.proxyConnections++
	utils.Proxy(conn1, conn2, time.Duration(t.MaxProxyLifetime)*time.Second)
	t.proxyConnections--
}
