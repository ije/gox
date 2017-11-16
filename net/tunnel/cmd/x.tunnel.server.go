package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ije/gox/config"
	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
)

func main() {
	cfile := flag.String("c", "/etc/x.tunnel/server.conf", "x.tunnel server runtime configuration file")
	debug := flag.Bool("d", false, "debug mode")
	flag.Parse()

	cfg, err := config.New(*cfile)
	if err != nil {
		fmt.Println("init config failed:", err)
		return
	}

	logger := &log.Logger{}
	if !*debug {
		logger.SetLevelByName("info")
	}
	tunnel.SetLogger(logger)

	ts := tunnel.Server{
		Port:     uint16(cfg.Int("port", 333)),
		HTTPPort: uint16(cfg.Int("http-port", 8080)),
		Secret:   cfg.String("password", ""),
	}

	for name, section := range cfg.ExtendedSections() {
		port := section.Int("port", 0)
		if port > 0 && port < 1<<16 && strings.HasPrefix(name, "tunnel:") {
			name = strings.TrimPrefix(name, "tunnel:")
			ts.AddTunnel(name, uint16(port), section.Int("maxClientConnections", 1))
			logger.Infof("tunnel(%s) with port(%d) added\n", name, port)
		}
	}

	logger.Infof("x.tunnel server started")
	ts.Serve()
}
