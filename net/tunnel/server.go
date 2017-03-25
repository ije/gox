package tunnel

import (
	"net"
	"sync"
	"time"

	"github.com/ije/gox/net/aestcp"
	"github.com/ije/gox/utils"
)

type Server struct {
	Port    uint16
	AESKey  string
	lock    sync.RWMutex
	tunnels map[string]*Tunnel
}

func (s *Server) AddTunnel(name string, port uint16, maxClientConnections int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

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
	l, err := aestcp.Listen("tcp", strf(":%d", s.Port), []byte(s.AESKey))
	if err != nil {
		return
	}

	return listen(l, s.handleConn)
}

func (s *Server) getTunnel(name string) (tunnel *Tunnel, ok bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if len(s.tunnels) > 0 {
		tunnel, ok = s.tunnels[name]
	}
	return
}

func (s *Server) handleConn(conn net.Conn) {
	tunnelName := make(chan string, 1)
	ec := make(chan error, 1)

	go func() {
		flag, data, err := parseMessage(conn)
		if err != nil {
			ec <- err
			return
		}

		if flag == "register" {
			var client Client
			err := utils.DecodeGobBytes(data, &client)
			if err == nil {
				if t, ok := s.getTunnel(client.TunnelName); ok {
					if t.Port == client.TunnelPort {
						err = sendMessage(conn, "registered", nil)
					} else {
						err = sendMessage(conn, "error", []byte(strf("tunnel has be registered with port %d", t.Port)))
					}
				} else {
					err = s.AddTunnel(client.TunnelName, client.TunnelPort, client.Connections)
					if err != nil {
						err = sendMessage(conn, "error", []byte(err.Error()))
					} else {
						err = sendMessage(conn, "registered", nil)
					}
				}
			}

			if err != nil {
				ec <- err
			} else {
				ec <- errf("registered")
			}
			return
		}

		if flag != "hello" {
			ec <- errf("invalid handshake message")
			return
		}

		tunnelName <- string(data)
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

	tunnel, ok := s.getTunnel(<-tunnelName)
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
		case <-tunnel.connQueue:
			_, err := conn.Write([]byte{1})
			if err != nil {
				tunnel.clientConns <- nil
				conn.Close()
			} else {
				tunnel.clientConns <- conn
				log.Debugf("tunnel(%s) client connection activated", tunnel.Name)
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
