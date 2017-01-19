package tunnel

import (
	"io"
	"net"
	"time"
)

type Service struct {
	Name        string
	Port        uint16
	clientConns chan net.Conn
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
	case clientConn = <-s.clientConns:
	case <-time.After(15 * time.Second):
		conn.Close()
		return
	}

	err := sendData(clientConn, "start-proxy", []byte(s.Name))
	if err != nil {
		if err == io.EOF || err.Error() == "use of closed network connection" {
			s.handleConn(conn)
			return
		}

		conn.Close()
		clientConn.Close()
		log.Warnf("x.tunnel service(%s): send data: %v", s.Name, err)
		return
	}

	log.Infof("x.tunnel server: service(%s) client connection activated (%d/%d)", s.Name, len(s.clientConns), cap(s.clientConns))
	proxy(conn, clientConn)
}
