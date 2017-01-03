package tunnel

import (
	"io"
	"net"
	"time"
)

type Service struct {
	Name       string
	Port       uint16
	clientConn net.Conn
}

func (s *Service) Serve() (err error) {
	l, err := net.Listen("tcp", strf(":%d", s.Port))
	if err != nil {
		return
	}

	go func(s *Service, l net.Listener) {
		defer l.Close()

		var tempDelay time.Duration
		for {
			conn, e := l.Accept()
			if e != nil {
				if ne, ok := e.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					time.Sleep(tempDelay)
					log.Warnf("net: Accept error: %v; retrying in %v", e, tempDelay)
					continue
				}
				log.Errorf("net: Accept error: %v", e)
				return
			}
			tempDelay = 0
			go s.handleConn(conn)
		}
	}(s, l)

	return
}

func (s *Service) handleConn(conn net.Conn) {
	defer conn.Close()

	if s.clientConn == nil {
		return
	}

	err := sendData(s.clientConn, "hello", []byte(s.Name))
	if err != nil {
		if err == io.EOF {
			s.clientConn = nil
		}
		return
	}

	closeChan := make(chan struct{}, 1)

	go func(clientConn net.Conn, conn net.Conn, Exit chan struct{}) {
		_, err := io.Copy(clientConn, conn)
		if err != nil && err == io.EOF {
			s.clientConn = nil
		}
		closeChan <- struct{}{}
	}(s.clientConn, conn, closeChan)

	go func(clientConn net.Conn, conn net.Conn, Exit chan struct{}) {
		io.Copy(conn, clientConn)
		closeChan <- struct{}{}
	}(s.clientConn, conn, closeChan)

	<-closeChan
}
