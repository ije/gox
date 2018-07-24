package tunnel

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/ije/gox/utils"
)

type Server struct {
	Port    uint16 // tunnel server port
	SSPort  uint16 // status server port
	tunnels map[string]*Tunnel
}

func (s *Server) AddTunnel(name string, port uint16, maxConnections int, maxLifetime int) error {
	if maxConnections <= 0 {
		maxConnections = 1
	}

	tunnel := &Tunnel{
		Name:             name,
		Port:             port,
		ProxyConnections: 0,
		ProxyMaxLifetime: maxLifetime,
		MaxConnections:   maxConnections,
		connQueue:        make(chan net.Conn, maxConnections),
		connPool:         make(chan net.Conn, maxConnections),
	}

	if s.tunnels == nil {
		s.tunnels = map[string]*Tunnel{}
	}
	s.tunnels[name] = tunnel

	return tunnel.Serve()
}

func (s *Server) Serve() (err error) {
	go func() {
		if s.SSPort > 0 {
			http.ListenAndServe(fmt.Sprintf(":%d", s.SSPort), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(s.tunnels)
				w.Header().Set("Content-Type", "application/json")
			}))
		}
	}()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
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
			err = fmt.Errorf("invalid handshake message")
			return
		}

		_, err = conn.Write([]byte{1})
		if err != nil {
			return
		}

		flag = f
		tunnelName = string(d)
		return
	}, 15*time.Second) != nil {
		conn.Close() // connection will be closed when can not get a valid handshake message and send a response in 15 seconds
		return
	}

	tunnel, ok := s.tunnels[tunnelName]
	if !ok {
		conn.Close()
		return
	}

	if flag == "proxy" {
		if len(tunnel.Client) == 0 {
			conn.Close()
			return
		}

		select {
		case c := <-tunnel.connPool:
			tunnel.proxy(conn, c)
		case <-time.After(5 * time.Second):
			conn.Close()
		}
		return
	}

	// only one tunnel client connection can be keep-alive
	if len(tunnel.Client) > 0 {
		conn.Close()
		return
	}

	remoteAddr, _ := utils.SplitByLastByte(conn.RemoteAddr().String(), ':')
	tunnel.activate(remoteAddr)
	defer func() {
		tunnel.Online = false
		tunnel.Client = ""
	}()

	for {
		select {
		case c := <-tunnel.connQueue:
			ret, err := exchangeByte(conn, 2, 5*time.Second)
			if err != nil {
				conn.Close()
				return
			}

			if ret == 1 {
				tunnel.activate(remoteAddr)
				tunnel.connPool <- c
				log.Debugf("tunnel(%s) proxy connection activated", tunnel.Name)
			} else {
				c.Close()
			}

		// heartbeat check
		case <-time.After(time.Second):
			ret, err := exchangeByte(conn, 1, 5*time.Second)
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
