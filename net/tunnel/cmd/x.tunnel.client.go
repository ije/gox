package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ije/gox/config"
	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
)

func main() {
	cfile := flag.String("c", "/etc/x.tunnel/client.conf", "x.tunnel client runtime configuration file")
	debug := flag.Bool("d", false, "debug mode")
	flag.Parse()

	cfg, err := config.New(*cfile)
	if err != nil {
		fmt.Println("load the configuration failed:", err)
		return
	}

	logger := &log.Logger{}
	if !*debug {
		logger.SetLevelByName("info")
	}
	tunnel.SetLogger(logger)

	ts := cfg.String("server", "")
	tsPassword := cfg.String("password", "")

	var clients int
	for name, section := range cfg.ExtendedSections() {
		port := section.Int("forward-port", 0)
		if port > 0 && port < 1<<16 && strings.HasPrefix(name, "tunnel:") {
			name = strings.TrimPrefix(name, "tunnel:")

			go func(server string, password string, name string, port uint16) {
				for {
					tc := &tunnel.Client{
						Server:      server,
						Password:    password,
						Tunnel:      name,
						ForwardPort: port,
						Connections: section.Int("connections", 1),
					}
					tc.Run()

					time.Sleep(time.Second)
				}
			}(ts, tsPassword, name, uint16(port))

			logger.Infof("tunnel %s added", name)
			clients++
		}
	}

	if clients > 0 {
		logger.Infof("x.tunnel client started")
		for {
			time.Sleep(time.Hour)
		}
	}
}
