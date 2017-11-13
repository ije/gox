package tunnel

import (
	"fmt"
	"io"
	"net"
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

	var tempDelay time.Duration
	for {
		conn, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if tempDelay > time.Second {
					tempDelay = time.Second
				}
				time.Sleep(tempDelay)
				log.Warnf("net: Accept error: %v; retrying in %v", e, tempDelay)
				continue
			}
			return e
		}
		tempDelay = 0
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
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func proxy(conn1 net.Conn, conn2 net.Conn) (err error) {
	if conn1 == nil || conn2 == nil {
		return errf("invalid arguments")
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
	conn1.Close()
	conn2.Close()
	return
}

func errf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func strf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
