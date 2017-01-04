package tunnel

import (
	"net"
	"time"
)

type Service struct {
	Name       string
	Port       uint16
	clientConn chan net.Conn
}

func (s *Service) Serve() (err error) {
	l, err := net.Listen("tcp", strf(":%d", s.Port))
	if err != nil {
		return
	}

	go listen(l, s.handleConn)
	return
}

func (s *Service) handleConn(conn net.Conn) {
	var clientConn net.Conn
	select {
	case clientConn = <-s.clientConn:
	case <-time.After(10 * time.Second):
		conn.Close()
		return
	}

	err := sendData(clientConn, "start-proxy", []byte(s.Name))
	if err != nil {
		log.Warnf("x.tunnel service(%s): send data: %v", s.Name, err)
		conn.Close()
		clientConn.Close()
		return
	}

	proxy(conn, clientConn)
}
