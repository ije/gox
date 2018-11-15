package tunnel

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Server struct {
	lock     sync.RWMutex
	Port     uint16 // tunnel service port
	HTTPPort uint16
	tunnels  map[string]*Tunnel
}

func (s *Server) ActivateTunnel(name string, port uint16, maxConnections int, maxProxyLifetime int) *Tunnel {
	s.lock.Lock()
	defer s.lock.Unlock()

	if maxConnections <= 0 {
		maxConnections = 1
	}

	if s.tunnels == nil {
		s.tunnels = map[string]*Tunnel{}
	} else if t, ok := s.tunnels[name]; ok {
		if t.Port == port {
			if t.MaxConnections < maxConnections {
				t.MaxConnections = maxConnections
				connQueue := make(chan net.Conn, maxConnections)
				connPool := make(chan net.Conn, maxConnections)
				close(t.connQueue)
				close(t.connPool)
				for conn := range t.connQueue {
					connQueue <- conn
				}
				for conn := range t.connPool {
					connPool <- conn
				}
				t.connQueue = connQueue
				t.connPool = connPool
			}
			if t.MaxProxyLifetime != maxProxyLifetime {
				t.MaxProxyLifetime = maxProxyLifetime
			}
			return t
		} else {
			t.close()
			delete(s.tunnels, name)
		}
	}

	tunnel := &Tunnel{
		Name:             name,
		Port:             port,
		MaxConnections:   maxConnections,
		MaxProxyLifetime: maxProxyLifetime,
		connQueue:        make(chan net.Conn, maxConnections),
		connPool:         make(chan net.Conn, maxConnections),
	}
	s.tunnels[name] = tunnel
	go tunnel.ListenAndServe()
	return tunnel
}

func (s *Server) Serve() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return
	}

	if s.HTTPPort > 0 {
		go s.serveHTTP()
	}
	return listen(l, s.handleConn)
}

func (s *Server) serveHTTP() (err error) {
	return http.ListenAndServe(fmt.Sprintf(":%d", s.HTTPPort), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wh := w.Header()
		wh.Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			wh.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
			wh.Set("Access-Control-Allow-Headers", "Accept,Accept-Encoding,Accept-Lang,Content-Type,Authorization,X-Requested-With")
			wh.Set("Access-Control-Allow-Credentials", "true")
			wh.Set("Access-Control-Max-Age", "60")
			w.WriteHeader(204)
			return
		}

		endpoint := strings.Trim(strings.TrimSpace(r.URL.Path), "/")
		if endpoint == "" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("x-tunnel-server"))
		} else if endpoint == "clients" {
			js := []map[string]interface{}{}
			s.lock.RLock()
			for _, t := range s.tunnels {
				meta := map[string]interface{}{
					"name":             t.Name,
					"port":             t.Port,
					"clientAddr":       t.clientAddr,
					"online":           t.online,
					"error":            t.err.Error(),
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
			s.lock.RUnlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(js)
		} else {
			http.Error(w, http.StatusText(404), 404)
		}
	}))
}

func (s *Server) handleConn(conn net.Conn) {
	var flag string
	var tunnel *Tunnel

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

		if flag == "hello" {
			var t Tunnel
			if gob.NewDecoder(bytes.NewReader(data)).Decode(&t) == nil {
				tunnel = s.ActivateTunnel(t.Name, t.Port, t.MaxConnections, t.MaxProxyLifetime)
				if tunnel.err != nil {
					err = fmt.Errorf("can not activate tunnel(%s)", t.Name)
					return
				}
			} else {
				err = fmt.Errorf("invalid hello message")
				return
			}
		} else if flag == "proxy" {
			var ok bool
			s.lock.RLock()
			tunnel, ok = s.tunnels[string(data)]
			s.lock.RUnlock()
			if !ok || tunnel.err != nil {
				err = fmt.Errorf("can not proxy tunnel(%s)", string(data))
				return
			}
		} else {
			err = fmt.Errorf("invalid flag")
			return
		}

		_, err = conn.Write([]byte{1})
		return
	}, 15*time.Second); err != nil {
		log.Warn("first touch:", err)
		conn.Close()
		return
	}

	if flag == "proxy" {
		tunnel.proxy(conn, <-tunnel.connPool)
		log.Debugf("server: tunnel(%s) start to proxy, current connPool has %d connections", tunnel.Name, len(tunnel.connPool))
		return
	}

	tunnel.activate(conn.RemoteAddr())
	defer tunnel.unactivate()

	log.Debugf("server: start to lookup connections from tunnel(%s)", tunnel.Name)
	for {
		select {
		case c := <-tunnel.connQueue:
			startTime := time.Now()
			ret, err := exchangeByte(conn, 2, 15*time.Second)
			if err != nil {
				conn.Close()
				c.Close()
				return
			}

			tunnel.activate(conn.RemoteAddr())
			if ret == 1 {
				tunnel.connPool <- c
				log.Debugf("server: tunnel(%s) is hit by proxy request, token %v", tunnel.Name, time.Now().Sub(startTime))
			} else if ret == 0 {
				c.Close()
			} else {
				conn.Close()
				return
			}

		// heart beat
		case <-time.After(10 * time.Second):
			startTime := time.Now()
			ret, err := exchangeByte(conn, 1, 15*time.Second)
			if err != nil || ret != 1 {
				conn.Close()
				return
			}

			tunnel.activate(conn.RemoteAddr())
			log.Debugf("server: tunnel(%s) is hit by heart beat, token %v", tunnel.Name, time.Now().Sub(startTime))
		}
	}
}
