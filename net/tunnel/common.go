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
		time.Sleep(100 * time.Millisecond)
	}
	return
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

func proxy(conn net.Conn, proxyConn net.Conn) {
	defer conn.Close()
	defer proxyConn.Close()

	closeChan := make(chan struct{}, 1)

	go func(conn net.Conn, proxyConn net.Conn, cc chan struct{}) {
		io.Copy(proxyConn, conn)
		cc <- struct{}{}
	}(proxyConn, conn, closeChan)

	go func(conn net.Conn, proxyConn net.Conn, cc chan struct{}) {
		io.Copy(conn, proxyConn)
		cc <- struct{}{}
	}(proxyConn, conn, closeChan)

	<-closeChan
}

func errf(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func strf(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}
