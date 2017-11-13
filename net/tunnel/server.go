package tunnel

import (
	"net"
	"time"

	"github.com/ije/gox/net/aestcp"
)

type Server struct {
	Port    uint16
	Secret  string
	tunnels map[string]*Tunnel
}

func (s *Server) AddTunnel(name string, port uint16, maxClientConnections int) error {
	if maxClientConnections <= 0 {
		maxClientConnections = 1
	}

	tunnel := &Tunnel{
		Name:      name,
		Port:      port,
		connQueue: make(chan net.Conn, maxClientConnections),
		connPool:  make(chan net.Conn, maxClientConnections),
	}

	if s.tunnels == nil {
		s.tunnels = map[string]*Tunnel{}
	}
	s.tunnels[name] = tunnel

	return tunnel.Serve()
}

func (s *Server) Serve() (err error) {
	l, err := aestcp.Listen("tcp", strf(":%d", s.Port), []byte(s.Secret))
	if err != nil {
		return
	}

	return listen(l, s.handleConn)
}

func (s *Server) handleConn(conn net.Conn) {
	if len(s.tunnels) == 0 {
		conn.Close()
		return
	}

	fc := make(chan string, 1)
	tc := make(chan string, 1)
	ec := make(chan error, 1)

	go func(fc chan string, tc chan string, ec chan error) {
		flag, data, err := parseMessage(conn)
		if err != nil {
			ec <- err
			return
		}

		if flag != "hello" && flag != "proxy" {
			ec <- errf("invalid handshake message")
			return
		}

		fc <- flag
		tc <- string(data)
		ec <- nil
	}(fc, tc, ec)

	select {
	case err := <-ec:
		if err != nil {
			conn.Close()
			return
		}
	case <-time.After(3 * time.Second):
		conn.Close() // connection will be closed when can not get the valid handshake message in 3 seconds
		return
	}

	tunnel, ok := s.tunnels[<-tc]
	if !ok {
		conn.Close()
		return
	}

	_, err := conn.Write([]byte{1})
	if err != nil {
		conn.Close()
		return
	}

	if <-fc == "proxy" {
		select {
		case tc := <-tunnel.connPool:
			proxy(tc, conn)
		case <-time.After(6 * time.Second):
			conn.Close()
		}
		return
	}

	for {
		select {
		case tc := <-tunnel.connQueue:
			_, err := conn.Write([]byte{2})
			if err != nil {
				conn.Close()
				return
			}
			if tc != nil {
				tunnel.connPool <- tc
				log.Debugf("tunnel(%s) connection activated", tunnel.Name)
			}
		case <-time.After(time.Second):
			_, err := conn.Write([]byte{1})
			if err != nil {
				conn.Close()
				return
			}
			log.Debugf("tunnel %s hearbeat", tunnel.Name)
		}
	}
}
