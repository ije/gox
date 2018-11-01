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

	var clients int
	for name, section := range cfg.ExtendedSections() {
		forwardPort := section.Int("forward_port", 0)
		tunnelPort := section.Int("tunnel_port", 0)
		if forwardPort > 0 && forwardPort < 1<<16 && tunnelPort > 0 && tunnelPort < 1<<16 && strings.HasPrefix(name, "tunnel:") {
			name = strings.TrimPrefix(name, "tunnel:")
			tc := &tunnel.Client{
				Server: ts,
				Tunnel: tunnel.Tunnel{
					Name:             name,
					Port:             uint16(tunnelPort),
					MaxConnections:   section.Int("max_connections", 1),
					MaxProxyLifetime: section.Int("max_proxy_lifetime", 0),
				},
				ForwardPort: uint16(forwardPort),
			}
			go tc.Connect()
			clients++
		}
	}

	if clients > 0 {
		utils.WaitExit(func(sig os.Signal) bool {
			return true
		})
	}
}
