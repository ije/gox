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

func (s *Server) AddService(name string, port uint16) error {
	if s.services == nil {
		s.services = map[string]*Service{}
	}

	service := &Service{
		Name: name,
		Port: port,
	}

	s.services[name] = service
	return service.Serve()
}

func (s *Server) Serve() (err error) {
	l, err := aestcp.Listen("tcp", strf(":%d", s.Port), []byte(s.AESKey))
	if err != nil {
		return
	}
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
				log.Errorf("asetcp: Accept error: %v; retrying in %v", e, tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
		go s.handleConn(conn)
	}
	return
}

func (s *Server) handleConn(conn net.Conn) {
	flag, data, err := parseData(conn)
	if err != nil {
		conn.Close()
		return
	}

	if flag != "hello" {
		conn.Close()
		return
	}

	if s.services == nil {
		conn.Close()
		return
	}

	serviceName := string(data)
	service, ok := s.services[serviceName]
	if !ok {
		conn.Close()
		return
	}

	service.clientConn = conn
	log.Debugf("server: service(%s) client connected", service.Name)
}
