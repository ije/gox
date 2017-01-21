package tunnel

import (
	"io"
	"net"
	"strings"
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

	go listen(l, func(conn net.Conn) {
		s.handleConn(-1, conn)
	})
	return
}

func (s *Service) handleConn(firstByte int, conn net.Conn) {
	var clientConn net.Conn
	select {
	case clientConn = <-s.clientConns:
	case <-time.After(15 * time.Second):
		conn.Close()
		return
	}

	buf := make([]byte, 1)
	if firstByte > -1 && firstByte < 256 {
		buf[0] = byte(firstByte)
	} else {
		_, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			go func(connChan chan net.Conn, conn net.Conn) {
				connChan <- conn
			}(s.clientConns, clientConn)
			return
		}
	}

	_, err := clientConn.Write(buf)
	if err != nil {
		if err == io.EOF || strings.HasSuffix(err.Error(), "use of closed network connection") { // EOF or net.OpError.Err is "use of closed network connection"
			s.handleConn(int(buf[0]), conn)
			return
		}

		conn.Close()
		clientConn.Close()
		log.Warnf("x.tunnel service(%s): send data: %v", s.Name, err)
		return
	}

	log.Debugf("service(%s) client connection activated (%d/%d)", s.Name, len(s.clientConns), cap(s.clientConns))
	proxy(conn, clientConn)
}
