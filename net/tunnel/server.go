package tunnel

import (
	"net"
	"time"

	"github.com/ije/gox/net/aestcp"
)

type Server struct {
	Port     uint16
	Password string
	tunnels  map[string]*Tunnel
}

func (s *Server) AddTunnel(name string, port uint16, maxClientConnections int) error {
	if s.tunnels == nil {
		s.tunnels = map[string]*Tunnel{}
	}

	if maxClientConnections <= 0 {
		maxClientConnections = 1
	}

	tunnel := &Tunnel{
		Name:        name,
		Port:        port,
		connQueue:   make(chan struct{}, maxClientConnections),
		clientConns: make(chan net.Conn, maxClientConnections),
	}

	s.tunnels[name] = tunnel
	return tunnel.Serve()
}

func (s *Server) Serve() (err error) {
	l, err := aestcp.Listen("tcp", strf(":%d", s.Port), []byte(s.Password))
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

	tunnelName := make(chan string, 1)
	ec := make(chan error, 1)
	proxy := false

	go func() {
		flag, data, err := parseMessage(conn)
		if err != nil {
			ec <- err
			return
		}

		if flag != "hello" && flag != "proxy" {
			ec <- errf("invalid handshake message")
			return
		}

		proxy = flag == "proxy"
		tunnelName <- string(data)
		ec <- nil
	}()

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

	tunnel, ok := s.tunnels[<-tunnelName]
	if !ok {
		conn.Close()
		return
	}

	_, err := conn.Write([]byte{1})
	if err != nil {
		conn.Close()
		return
	}

	if proxy {
		tunnel.clientConns <- conn
		log.Debugf("tunnel(%s) client connection activated", tunnel.Name)
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
		case <-tunnel.connQueue:
			_, err := conn.Write([]byte{1})
			if err != nil {
				conn.Close()
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
