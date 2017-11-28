package tunnel

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	logger "github.com/ije/gox/log"
	"github.com/ije/gox/net/aestcp"
)

var log = &logger.Logger{}

func init() {
	log.SetLevel(logger.L_INFO)
}

func SetLogger(l *logger.Logger) {
	if l != nil {
		log = l
	}
}

func SetLogLevel(level string) {
	log.SetLevelByName(level)
}

func listen(l net.Listener, connHandler func(net.Conn)) error {
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
			return err
		}
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			err := tcpConn.SetKeepAlive(true)
			if err != nil {
				return err
			}
		}
		go connHandler(conn)
	}
}

func dial(network string, address string, aes string) (conn net.Conn, err error) {
	for i := 0; i < 3; i++ {
		if len(aes) > 0 {
			conn, err = aestcp.Dial(network, address, []byte(aes))
		} else {
			conn, err = net.Dial(network, address)
		}
		if err == nil {
			return
		}
		time.Sleep(time.Second / 2)
	}
	return
}

func proxy(conn1 net.Conn, conn2 net.Conn) (err error) {
	if conn1 == nil || conn2 == nil {
		return errf("invalid connections")
	}

	ec := make(chan error, 1)

	go func(conn1 net.Conn, conn2 net.Conn, ec chan error) {
		_, err := io.Copy(conn1, conn2)
		ec <- err
	}(conn1, conn2, ec)

	go func(conn1 net.Conn, conn2 net.Conn, ec chan error) {
		_, err := io.Copy(conn2, conn1)
		ec <- err
	}(conn1, conn2, ec)

	err = <-ec
	go conn1.Close()
	go conn2.Close()
	return
}

// exchange 1 byte data with timeout
func exchangeByte(conn net.Conn, b byte, timeout time.Duration) (ret byte, err error) {
	ec := make(chan error, 1)
	bc := make(chan byte, 1)
	go func(conn net.Conn, bc chan byte, ec chan error) {
		_, err := conn.Write([]byte{b})
		if err != nil {
			ec <- err
			return
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err != nil {
			ec <- err
			return
		}

		bc <- buf[0]
		ec <- nil
	}(conn, bc, ec)

	if timeout <= 0 {
		err = <-ec
		if err == nil {
			ret = <-bc
		}
		return
	}

	select {
	case err = <-ec:
		if err == nil {
			ret = <-bc
		}
	case <-time.After(timeout):
		err = errf("timeout")
	}
	return
}

func errf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func strf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
