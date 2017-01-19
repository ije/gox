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

	var serviceName string
	ec := make(chan error, 1)

	go func() {
		flag, data, err := parseData(conn)
		if err != nil {
			ec <- err
			return
		}

		if flag != "hello" {
			ec <- errf("invalid handshake message")
			return
		}

		serviceName = string(data)
		ec <- nil
	}()

	// connection will be closed when can not get the right handshake message in 5 seconds
	select {
	case err := <-ec:
		if err != nil {
			conn.Close()
			return
		}
	case <-time.After(5 * time.Second):
		conn.Close()
		return
	}

	service, ok := s.services[serviceName]
	if !ok {
		conn.Close()
		return
	}

	l, c := len(service.clientConns), cap(service.clientConns)
	if l == c {
		(<-service.clientConns).Close()
		l--
	}

	service.clientConns <- conn
	log.Infof("x.tunnel server: service(%s) client connected (%d/%d)", service.Name, l+1, c)
}
