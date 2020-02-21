package tunnel

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

const heartBeatInterval = 15

type Server struct {
	lock    sync.RWMutex
	Port    uint16 // tunnel service port
	tunnels map[string]*Tunnel
}

func (s *Server) ActivateTunnel(name string, port uint16, maxProxyLifetime int) *Tunnel {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.tunnels == nil {
		s.tunnels = map[string]*Tunnel{}
	} else if t, ok := s.tunnels[name]; ok {
		if t.Port == port {
			t.MaxProxyLifetime = maxProxyLifetime
			return t
		}
		t.close()
		delete(s.tunnels, name)
	}

	tunnel := &Tunnel{
		Name:             name,
		Port:             port,
		MaxProxyLifetime: maxProxyLifetime,
		connQueue:        make(chan net.Conn, 1000),
		connPool:         make(chan net.Conn, 1000),
	}
	s.tunnels[name] = tunnel
	go func(t *Tunnel) {
		for {
			t.ListenAndServe()
		}
	}(tunnel)
	return tunnel
}

func (s *Server) Serve() (err error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return
	}

	return listen(l, s.handleConn)
}



func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tunnels := TunnelSlice{}
	s.lock.RLock()
	for _, t := range s.tunnels {
		tunnels = append(tunnels, TunnelInfo{
			Name:             t.Name,
			Port:             t.Port,
			MaxProxyLifetime: t.MaxProxyLifetime,
			ClientAddr:       t.clientAddr,
			Online:           t.online,
			ProxyConnections: t.proxyConnections,
			ConnPoolLength:   len(t.connPool),
			ConnQueueLength:  len(t.connQueue),
		})
	}
	s.lock.RUnlock()

	sort.Sort(tunnels)

	w.Header().Set("Content-Type", "application/json")
	je := json.NewEncoder(w)
	je.SetIndent("", "\t")
	je.Encode(map[string]interface{}{
		"port":    s.Port,
		"tunnels": tunnels,
	})
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
				tunnel = s.ActivateTunnel(t.Name, t.Port, t.MaxProxyLifetime)
			} else {
				err = fmt.Errorf("invalid hello message")
				return
			}
		} else if flag == "proxy" {
			var ok bool
			s.lock.RLock()
			tunnel, ok = s.tunnels[string(data)]
			s.lock.RUnlock()
			if !ok {
				err = fmt.Errorf("can not proxy tunnel(%s)", string(data))
				return
			}
		} else {
			err = fmt.Errorf("invalid flag")
			return
		}

		_, err = conn.Write([]byte{1})
		return
	}, 2*heartBeatInterval*time.Second); err != nil {
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
			ret, err := exchangeByte(conn, 2, 2*heartBeatInterval*time.Second)
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
		case <-time.After(heartBeatInterval * time.Second):
			startTime := time.Now()
			ret, err := exchangeByte(conn, 1, 2*heartBeatInterval*time.Second)
			if err != nil || ret != 1 {
				conn.Close()
				return
			}

			tunnel.activate(conn.RemoteAddr())
			log.Debugf("server: tunnel(%s) is hit by heart beat, token %v", tunnel.Name, time.Now().Sub(startTime))
		}
	}
}
