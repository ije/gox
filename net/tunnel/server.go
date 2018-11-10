package tunnel

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ije/gox/utils"
)

type Server struct {
	Port     uint16 // tunnel service port
	HTTPPort uint16
	tunnels  map[string]*Tunnel
}

func (s *Server) AddTunnel(name string, port uint16, maxConnections int, maxProxyLifetime int) error {
	if maxConnections <= 0 {
		maxConnections = 1
	}

	tunnel := &Tunnel{
		Name:             name,
		Port:             port,
		MaxConnections:   maxConnections,
		MaxProxyLifetime: maxProxyLifetime,
		connQueue:        make(chan net.Conn, maxConnections),
		connPool:         make(chan net.Conn, maxConnections),
	}

	if s.tunnels == nil {
		s.tunnels = map[string]*Tunnel{}
	} else if t, ok := s.tunnels[name]; ok && t.Port == port {
		if t.MaxConnections != maxConnections {
			t.MaxConnections = maxConnections
			t.connQueue = make(chan net.Conn, maxConnections)
			t.connPool = make(chan net.Conn, maxConnections)
		}
		t.MaxProxyLifetime = maxProxyLifetime
		return nil
	}

	err := tunnel.Serve()
	if err == nil {
		s.tunnels[name] = tunnel
	}
	return err
}

func (s *Server) Serve() (err error) {
	go func() {
		if s.HTTPPort > 0 {
			http.ListenAndServe(fmt.Sprintf(":%d", s.HTTPPort), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				endpoint := strings.Trim(strings.TrimSpace(r.URL.Path), "/")
				if endpoint == "" {
					w.Header().Set("Content-Type", "text/plain")
					w.Write([]byte("tunnel-x-server"))
				} else if endpoint == "clients" {
					js := []map[string]interface{}{}
					for _, t := range s.tunnels {
						meta := map[string]interface{}{
							"name":             t.Name,
							"port":             t.Port,
							"client":           t.client,
							"online":           t.online,
							"maxConnections":   t.MaxConnections,
							"proxyConnections": t.proxyConnections,
							"connPoolLength":   len(t.connPool),
							"connQueueLength":  len(t.connQueue),
						}
						if t.MaxProxyLifetime > 0 {
							meta["maxProxyLifetime"] = t.MaxProxyLifetime
						}
						js = append(js, meta)
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(js)
				} else {
					http.Error(w, http.StatusText(400), 400)
				}
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

	if err := dotimeout(func() (err error) {
		var data []byte
		flag, data, err = parseMessage(conn)
		if err != nil {
			return
		}

		if flag != "hello" && flag != "proxy" {
			err = fmt.Errorf("invalid handshake message")
			return
		}

		var ret byte = 0
		if flag == "hello" {
			var tunnel Tunnel
			if gob.NewDecoder(bytes.NewReader(data)).Decode(&tunnel) == nil {
				if s.AddTunnel(tunnel.Name, tunnel.Port, tunnel.MaxConnections, tunnel.MaxProxyLifetime) == nil {
					tunnelName = tunnel.Name
					ret = 1
				}
			}
		} else if flag == "proxy" {
			tunnelName = string(data)
			if _, ok := s.tunnels[tunnelName]; ok {
				ret = 1
			}
		}

		_, err = conn.Write([]byte{ret})
		if err != nil {
			return
		}

		return
	}, 10*time.Second); err != nil {
		log.Warn("first touch:", err)
		conn.Close() // connection will be closed when can not get a valid handshake message and send a response in 10 seconds
		return
	}

	tunnel, ok := s.tunnels[tunnelName]
	if !ok {
		log.Warn("bad tunnel name:", tunnelName)
		conn.Close()
		return
	}

	if flag == "proxy" {
		tunnel.proxy(conn, <-tunnel.connPool)
		return
	}

	remoteAddr, _ := utils.SplitByLastByte(conn.RemoteAddr().String(), ':')
	tunnel.activate(remoteAddr)

	for {
		select {
		case c := <-tunnel.connQueue:
			ret, err := exchangeByte(conn, 2, 15*time.Second)
			if err != nil {
				conn.Close()
				c.Close()
				return
			}

			tunnel.activate(remoteAddr)
			if ret == 1 {
				tunnel.connPool <- c
				log.Debugf("tunnel(%s) proxy connection activated", tunnel.Name)
			} else {
				c.Close()
			}

		// heartbeat check
		case <-time.After(10 * time.Second):
			ret, err := exchangeByte(conn, 1, 15*time.Second)
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
