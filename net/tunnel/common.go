package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	logger "github.com/ije/gox/log"
	"github.com/ije/gox/net/aestcp"
)

var log = &logger.Logger{}

func SetLogLevel(level string) {
	log.SetLevelByName(level)
}

func SetLogQuite(quite bool) {
	log.SetQuite(quite)
}

var XTunnelHead = []byte{'X', '-', 'T', 'U', 'N', 'N', 'E', 'L'}

func errf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func strf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func dial(network string, address string, aes string) (conn net.Conn, err error) {
	for i := 0; i < 10; i++ {
		if len(aes) > 0 {
			conn, err = aestcp.Dial(network, address, []byte(aes))
		} else {
			conn, err = net.Dial(network, address)
		}
		if err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	return
}

func sendData(conn net.Conn, flag string, data []byte) (err error) {
	flagLen := len(flag)
	if flagLen > 200 {
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

func parseData(conn net.Conn) (flag string, data []byte, err error) {
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
