package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func sendMessage(conn net.Conn, flag Flag, data []byte) (err error) {
	buf := bytes.NewBuffer(nil)
	buf.WriteByte(1)
	buf.WriteByte(byte(flag))

	dl := uint32(len(data))
	if dl > 0 {
		len := make([]byte, 4)
		binary.LittleEndian.PutUint32(len, dl)
		buf.WriteByte(1)
		buf.Write(len)
		buf.Write(data)
	} else {
		buf.WriteByte(0)
	}

	_, err = io.Copy(conn, buf)
	return
}

func parseMessage(conn net.Conn) (flag Flag, data []byte, err error) {
	// check head
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	if buf[0] != 1 {
		err = fmt.Errorf("invalid head")
		return
	}

	// parse flag
	buf = make([]byte, 2)
	_, err = conn.Read(buf)
	if err != nil {
		return
	}
	flag = Flag(buf[0])

	// parse data
	hasData := buf[1] == 1
	if hasData {
		buf = make([]byte, 4)
		_, err = conn.Read(buf)
		if err != nil {
			return
		}
		len := binary.LittleEndian.Uint32(buf)
		buf := bytes.NewBuffer(nil)
		_, err = io.CopyN(buf, conn, int64(len))
		if err != nil {
			return
		}
		data = buf.Bytes()
	}
	return
}
