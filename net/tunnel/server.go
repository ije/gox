package tunnel

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
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
	Port     uint16 // tunnel service port
	Password string
	passhash []byte
	lock     sync.RWMutex
	tunnels  map[string]*Tunnel
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
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(time.Minute)
		}

		go s.handleConn(conn)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var names sort.StringSlice
	s.lock.RLock()
	for _, t := range s.tunnels {
		names = append(names, t.Name)
	}
	s.lock.RUnlock()
	names.Sort()

	var tunnels []interface{}
	for _, name := range names {
		s.lock.RLock()
		t, ok := s.tunnels[name]
		s.lock.RUnlock()
		if ok {
			info := map[string]interface{}{
				"name":       t.Name,
				"port":       t.Port,
				"clientAddr": t.clientAddr,
				"online":     t.online,
				"listener":   nil,
			}
			if t.MaxProxyLifetime > 0 {
				info["maxProxyLifetime"] = t.MaxProxyLifetime
			}
			if t.listener != nil {
				info["listener"] = "ok"
			}
			tunnels = append(tunnels, info)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	jw := json.NewEncoder(w)
	jw.SetIndent("", "\t")
	jw.Encode(map[string]interface{}{
		"port":    s.Port,
		"tunnels": tunnels,
	})
}

func (s *Server) secret() []byte {
	if len(s.passhash) == sha1.Size {
		return s.passhash
	}

	s.passhash = genSecret(s.Password)
	return s.passhash
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	var tunnel *Tunnel

	flag, data, err := parseMessage(conn)
	if err != nil || len(data) < 20 || !bytes.Equal(s.secret(), data[:20]) {
		return
	}
	data = data[20:]

	if flag == FlagHello {
		dl := len(data)
		if dl > 0 && dl == 1+int(data[0])+2+4 {
			nl := int(data[0])
			name := string(data[1 : 1+nl])
			port := binary.LittleEndian.Uint16(data[1+nl:])
			maxProxyLifetime := binary.LittleEndian.Uint32(data[1+nl+2:])
			tunnel = s.activateTunnel(name, port, maxProxyLifetime)
			log.Printf("tunnel(%s) activated, port: %d, maxProxyLifetime: %ds", name, port, maxProxyLifetime)
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
		// unsupport flag
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

			flag, data, err := parseMessage(conn)
			if err != nil {
				c.Close()
				return
			}

			if flag != FlagReady {
				if flag == FlagError {
					log.Printf("tunnel(%s) client returns an error: %s", tunnel.Name, string(data))
				}
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

			flag, _, err := parseMessage(conn)
			if err != nil || flag != FlagHello {
				return
			}

			tunnel.activate(conn.RemoteAddr())
		}
	}
}

func (s *Server) activateTunnel(name string, port uint16, maxProxyLifetime uint32) *Tunnel {
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
		TunnelProps: &TunnelProps{
			Name:             name,
			Port:             port,
			MaxProxyLifetime: maxProxyLifetime,
		},
		crtime:    time.Now().Unix(),
		connQueue: make(chan net.Conn, 1000),
		connPool:  make(chan net.Conn, 1000),
	}
	s.tunnels[name] = tunnel
	go tunnel.ListenAndServe()
	return tunnel
}
