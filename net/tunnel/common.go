package tunnel

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	logger "github.com/ije/gox/log"
)

var errTimeout = fmt.Errorf("timeout")

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
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return err
			}
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
		}

		go connHandler(conn)
	}
}

func dial(network string, address string) (conn net.Conn, err error) {
	for i := 0; i < 3; i++ {
		conn, err = net.Dial(network, address)
		if err == nil {
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				tcpConn.SetKeepAlive(true)
			}
			return
		}
	}
	return
}

func proxy(conn1 net.Conn, conn2 net.Conn, timeout time.Duration) (err error) {
	if conn1 == nil || conn2 == nil {
		return fmt.Errorf("invalid connections")
	}

	ec1 := make(chan error, 1)
	ec2 := make(chan error, 1)

	go func(conn1 net.Conn, conn2 net.Conn, ec chan error) {
		_, err := io.Copy(conn1, conn2)
		ec <- err
	}(conn1, conn2, ec1)

	go func(conn1 net.Conn, conn2 net.Conn, ec chan error) {
		_, err := io.Copy(conn2, conn1)
		ec <- err
	}(conn1, conn2, ec2)

	if timeout > 0 {
		select {
		case err = <-ec1:
		case err = <-ec2:
		case <-time.After(timeout):
			err = errTimeout
		}
	} else {
		select {
		case err = <-ec1:
		case err = <-ec2:
		}
	}

	conn1.Close()
	conn2.Close()
	return
}

// exchange 1 byte data with timeout
func exchangeByte(conn net.Conn, b byte, timeout time.Duration) (ret byte, err error) {
	err = dotimeout(func() (err error) {
		_, err = conn.Write([]byte{b})
		if err != nil {
			return
		}

		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err != nil {
			return
		}

		ret = buf[0]
		return
	}, timeout)
	return
}

func dotimeout(handle func() error, timeout time.Duration) (err error) {
	var ec = make(chan error, 1)
	go func(ec chan error) {
		ec <- handle()
	}(ec)

	select {
	case err = <-ec:
	case <-time.After(timeout):
		err = errTimeout
	}
	return
}
