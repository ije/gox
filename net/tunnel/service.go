package tunnel

import (
	"net"
	"time"
)

type Service struct {
	Name        string
	Port        uint16
	connQueue   chan struct{}
	clientConns chan net.Conn
}

func (s *Service) Serve() (err error) {
	l, err := net.Listen("tcp", strf(":%d", s.Port))
	if err != nil {
		return
	}

	go listen(l, func(conn net.Conn) {
		s.handleConn(conn)
	})
	return
}

func (s *Service) handleConn(conn net.Conn) {
	s.connQueue <- struct{}{}

	var clientConn net.Conn
	select {
	case clientConn = <-s.clientConns:
		if clientConn == nil {
			conn.Close()
			return
		}
	case <-time.After(6 * time.Second):
		<-s.connQueue
		conn.Close()
		return
	}

	proxy(conn, clientConn)
}
