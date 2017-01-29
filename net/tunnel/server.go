package tunnel

import (
	"net"
	"time"

	"github.com/ije/gox/net/aestcp"
)

type Server struct {
	Port     uint16
	AESKey   string
	services map[string]*Service
}

func (s *Server) AddService(name string, port uint16, maxClientConnections int) error {
	if s.services == nil {
		s.services = map[string]*Service{}
	}

	if maxClientConnections <= 0 {
		maxClientConnections = 1
	}

	service := &Service{
		Name:        name,
		Port:        port,
		connQueue:   make(chan struct{}, maxClientConnections),
		clientConns: make(chan net.Conn, maxClientConnections),
	}

	s.services[name] = service
	return service.Serve()
}

func (s *Server) Serve() (err error) {
	l, err := aestcp.Listen("tcp", strf(":%d", s.Port), []byte(s.AESKey))
	if err != nil {
		return
	}

	return listen(l, s.handleConn)
}

func (s *Server) handleConn(conn net.Conn) {
	if len(s.services) == 0 {
		conn.Close()
		return
	}

	serviceName := make(chan string, 1)
	ec := make(chan error, 1)

	go func() {
		flag, data, err := parseMessage(conn)
		if err != nil {
			ec <- err
			return
		}

		if flag != "hello" {
			ec <- errf("invalid handshake message")
			return
		}

		serviceName <- string(data)
		ec <- nil
	}()

	// connection will be closed when can not get the valid handshake message in 3 seconds
	select {
	case err := <-ec:
		if err != nil {
			conn.Close()
			return
		}
	case <-time.After(3 * time.Second):
		conn.Close()
		return
	}

	service, ok := s.services[<-serviceName]
	if !ok {
		conn.Close()
		return
	}

	_, err := conn.Write([]byte{1})
	if err != nil {
		conn.Close()
		return
	}

	for {
		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err != nil || buf[0] != '!' {
			conn.Close()
			return
		}

		select {
		case <-service.connQueue:
			_, err := conn.Write([]byte{1})
			if err != nil {
				service.clientConns <- nil
				conn.Close()
			} else {
				service.clientConns <- conn
				log.Debugf("service(%s) client connection activated", service.Name)
			}
			return
		case <-time.After(time.Second):
			_, err := conn.Write([]byte{0})
			if err != nil {
				conn.Close()
				return
			}
		}
	}
}
