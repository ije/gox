package tunnel

import (
	"net"
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
	clientConn := <-s.clientConn
	err := sendData(clientConn, "start proxy", []byte(s.Name))
	if err != nil {
		conn.Close()
		return
	}

	proxy(conn, clientConn)
}
