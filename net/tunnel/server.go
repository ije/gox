package tunnel

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"
)

var heartBeatInterval = 15

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
			if t.MaxProxyLifetime != maxProxyLifetime {
				t.MaxProxyLifetime = maxProxyLifetime
			}
			return t
		}
		t.close()
		delete(s.tunnels, name)
	}

	tunnel := &Tunnel{
		Name:             name,
		Port:             port,
		MaxProxyLifetime: maxProxyLifetime,
		crtime:           time.Now().Unix(),
		connQueue:        make(chan net.Conn, 1000),
		connPool:         make(chan net.Conn, 1000),
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
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			tcpConn.SetKeepAlive(false)
		}

		go s.handleConn(conn)
	}
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
		})
	}
	s.lock.RUnlock()

	sort.Sort(tunnels)

	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)
	j.SetIndent("", "\t")
	j.Encode(map[string]interface{}{
		"port":    s.Port,
		"tunnels": tunnels,
	})
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	flag, data, err := parseMessage(conn)
	if err != nil {
		return
	}

	var tunnel *Tunnel

	if flag == FlagHello {
		var t TunnelInfo
		if gob.NewDecoder(bytes.NewReader(data)).Decode(&t) == nil {
			tunnel = s.ActivateTunnel(t.Name, t.Port, t.MaxProxyLifetime)
			log.Printf("tunnel(%s) activated, port: %d, maxProxyLifetime: %ds", t.Name, t.Port, t.MaxProxyLifetime)
		} else {
			return
		}
	} else if flag == FlagProxy {
		var ok bool
		s.lock.RLock()
		tunnel, ok = s.tunnels[string(data)]
		s.lock.RUnlock()
		if ok {
			tunnel.proxy(conn, <-tunnel.connPool)
		}
		return
	} else {
		// invalid flag
		return
	}

	tunnel.activate(conn.RemoteAddr())
	defer tunnel.unactivate()

	for {
		select {
		case c := <-tunnel.connQueue:
			err := sendMessage(conn, FlagProxy, nil)
			if err != nil {
				c.Close()
				return
			}

			tunnel.activate(conn.RemoteAddr())
			tunnel.connPool <- c

		// heart beat
		case <-time.After(time.Duration(heartBeatInterval) * time.Second):
			err := sendMessage(conn, FlagHello, nil)
			if err != nil {
				return
			}

			tunnel.activate(conn.RemoteAddr())
		}
	}
}
