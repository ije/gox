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

func (s *Server) AddTunnel(name string, port uint16, maxConnections int) error {
	if maxConnections <= 0 {
		maxConnections = 1
	}

	tunnel := &Tunnel{
		Name:      name,
		Port:      port,
		connQueue: make(chan net.Conn, maxConnections),
		connPool:  make(chan net.Conn, maxConnections),
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
	var flag string
	var tunnelName string

	if dotimeout(func() (err error) {
		f, d, err := parseMessage(conn)
		if err != nil {
			return
		}

		if f != "hello" && f != "proxy" {
			err = errf("invalid handshake message")
			return
		}

		_, err = conn.Write([]byte{1})
		if err != nil {
			return
		}

		flag = f
		tunnelName = string(d)
		return
	}, 3*time.Second) != nil {
		conn.Close() // connection will be closed when can not get a valid handshake message and send a response in 3 seconds
		return
	}

	tunnel, ok := s.tunnels[tunnelName]
	if !ok {
		conn.Close()
		return
	}

	remoteAddr, _ := utils.SplitByLastByte(conn.RemoteAddr().String(), ':')
	if flag == "proxy" {
		if len(tunnel.Client) == 0 || tunnel.Client != remoteAddr {
			conn.Close()
			return
		}

		select {
		case c := <-tunnel.connPool:
			proxy(conn, c, 0)
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
	defer func() {
		tunnel.Online = false
		tunnel.Client = ""
	}()

	conn.SetDeadline(time.Time{})
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
}
