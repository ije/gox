package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ije/gox/config"
	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
	"github.com/ije/gox/utils"
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

			tc := &tunnel.Client{
				Server:      ts,
				Password:    tsPassword,
				Tunnel:      name,
				ForwardPort: uint16(port),
				Connections: section.Int("connections", 1),
			}
			go tc.Run()

			logger.Infof("tunnel %s added", name)
			clients++
		}
	}

	if clients > 0 {
		logger.Infof("x.tunnel client started")
		utils.WaitExit(func(sig os.Signal) bool {
			return true
		})
	}
}
