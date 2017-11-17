package tunnel

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/ije/gox/net/aestcp"
	"github.com/ije/gox/utils"
)

type Server struct {
	Port     uint16
	HTTPPort uint16
	Secret   string
	tunnels  map[string]*Tunnel
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
	go func() {
		if s.HTTPPort > 0 {
			http.ListenAndServe(strf(":%d", s.HTTPPort), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(s.tunnels)
				w.Header().Set("Content-Type", "application/json")
			}))
		}
	}()

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

	go func(conn net.Conn, fc chan string, tc chan string, ec chan error) {
		flag, data, err := parseMessage(conn)
		if err != nil {
			ec <- err
			return
		}

		if flag != "hello" && flag != "proxy" {
			ec <- errf("invalid handshake message")
			return
		}

		_, err = conn.Write([]byte{1})
		if err != nil {
			ec <- err
			return
		}

		fc <- flag
		tc <- string(data)
		ec <- nil
	}(conn, fc, tc, ec)

	select {
	case err := <-ec:
		if err != nil {
			conn.Close()
			return
		}
	case <-time.After(3 * time.Second):
		conn.Close() // connection will be closed when can not get a valid handshake message and send a response in 3 seconds
		return
	}

	remoteAddr, _ := utils.SplitByLastByte(conn.RemoteAddr().String(), ':')

	tunnel, ok := s.tunnels[<-tc]
	if !ok {
		conn.Close()
		return
	}

	if <-fc == "proxy" {
		if len(tunnel.Client) == 0 || tunnel.Client != remoteAddr {
			conn.Close()
			return
		}

		select {
		case c := <-tunnel.connPool:
			proxy(conn, c)
		case <-time.After(6 * time.Second):
			conn.Close()
		}
		return
	}

	// only on birdge connection can be keep-alive
	if len(tunnel.Client) > 0 {
		conn.Close()
		return
	}

	tunnel.activate(remoteAddr)

	for {
		select {
		case c := <-tunnel.connQueue:
			ret, err := exchangeByte(conn, 2, 3*time.Second)
			if err != nil {
				conn.Close()
				return
			}

			if ret == 1 {
				tunnel.activate(remoteAddr)
				tunnel.connPool <- c
				log.Debugf("tunnel(%s) connection activated", tunnel.Name)
			} else {
				c.Close()
			}

		case <-time.After(time.Second):
			ret, err := exchangeByte(conn, 1, 3*time.Second)
			if err != nil {
				conn.Close()
				return
			}

			if ret == 1 {
				tunnel.activate(remoteAddr)
				log.Debugf("tunnel(%s) activated(heartbeat)", tunnel.Name)
			}
		}
	}

	tunnel.Online = false
	tunnel.Client = ""
}
