package main

import (
	"flag"

	"github.com/ije/gox/config"
	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
)

func main() {
	var cfile string
	var port int
	var password string
	var debug bool

	flag.StringVar(&cfile, "c", "/etc/x.tunnel/server.conf", "x.tunnel server runtime configuration file")
	flag.IntVar(&port, "port", 333, "server port")
	flag.StringVar(&password, "password", "", "password")
	flag.BoolVar(&debug, "d", false, "print debug message")
	flag.Parse()

	cfg, err := config.New(cfile)
	if err == nil {
		pw := cfg.String("password", "")
		if len(pw) > 0 {
			password = pw
		}
		p := cfg.Int("port", 0)
		if p > 0 {
			port = p
		}
	}

	logger := &log.Logger{}
	if !debug {
		logger.SetLevelByName("info")
	}
	tunnel.SetLogger(logger)

	ts := tunnel.Server{
		Port:   uint16(port),
		AESKey: password,
	}
	logger.Infof("x.tunnel server started")
	ts.Serve()
}
