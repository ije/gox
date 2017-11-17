package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

var XTunnelHead = []byte("X-TUNNEL")

type Tunnel struct {
	Name      string `json:"name"`
	Port      uint16 `json:"port"`
	Online    bool   `json:"online"`
	Client    string `json:"client,omitempty"`
	olTimer   *time.Timer
	connQueue chan net.Conn
	connPool  chan net.Conn
}

func (t *Tunnel) Serve() (err error) {
	l, err := net.Listen("tcp", strf(":%d", t.Port))
	if err != nil {
		return
	}

	go listen(l, func(conn net.Conn) {
		log.Debugf("tunnel(%s, Online:%v) new connection ", t.Name, t.Online)

		if !t.Online {
			conn.Close()
			return
		}

		t.connQueue <- conn
	})

	return
}

func (t *Tunnel) activate(remoteAddr string) {
	if t.olTimer != nil {
		t.olTimer.Stop()
	}
	t.olTimer = time.AfterFunc(3*time.Second, func() {
		t.olTimer = nil
		t.Online = false
		t.Client = ""
	})

	t.Online = true
	t.Client = remoteAddr
}

func sendMessage(conn net.Conn, flag string, data []byte) (err error) {
	flagLen := len(flag)
	if flagLen > 255 {
		err = fmt.Errorf("invalid flag")
		return
	}

	_, err = conn.Write(XTunnelHead)
	if err != nil {
		return
	}

	_, err = conn.Write([]byte{byte(flagLen)})
	if err != nil {
		return
	}
	_, err = conn.Write([]byte(flag))
	if err != nil {
		return
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(data)))
	_, err = conn.Write(buf)
	if err != nil {
		return
	}

	if len(data) > 0 {
		_, err = io.Copy(conn, bytes.NewReader(data))
	}
	return
}

func parseMessage(conn net.Conn) (flag string, data []byte, err error) {
	buf := make([]byte, 8)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	if !bytes.Equal(buf, XTunnelHead) {
		err = fmt.Errorf("invalid x-tunnel head")
		return
	}

	buf = make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	buf = make([]byte, int(buf[0]))
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	flag = string(buf)

	buf = make([]byte, 4)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	if dl := binary.LittleEndian.Uint32(buf); dl > 0 {
		buf := bytes.NewBuffer(nil)
		_, err = io.CopyN(buf, conn, int64(dl))
		if err == nil {
			data = buf.Bytes()
		}
	}
	return
}
