package utils

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
)

// WaitForExitSignal waits for the exit signal.
func WaitForExitSignal(callback func(os.Signal) bool) {
	if callback == nil {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP)
	<-c
}

// GetLocalIPList return the list of local ip address.
func GetLocalIPList() (list []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				list = append(list, ipnet.IP.String())
			}
		}
	}

	if len(list) == 0 {
		err = errors.New("not found")
	}
	return
}
