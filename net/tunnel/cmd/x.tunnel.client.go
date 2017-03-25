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
	cfile := flag.String("c", "/etc/x.tunnel/client.conf", "x.tunnel client runtime configuration file")
	debug := flag.Bool("d", false, "print mode message")
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

	tunnelServer := cfg.String("server", "")
	tunnelServerPassword := cfg.String("password", "")

	var clients []*tunnel.Client
	for key, section := range cfg.ExtendedSections() {
		tunnelPort := section.Int("tunnel-port", 0)
		localPort := section.Int("local-port", 0)
		if tunnelPort > 0 && tunnelPort < 1<<16 && localPort > 0 && localPort < 1<<16 && strings.HasPrefix(key, "tunnel:") {
			client := &tunnel.Client{
				Server:      tunnelServer,
				AESKey:      tunnelServerPassword,
				TunnelName:  strings.TrimPrefix(key, "tunnel:"),
				TunnelPort:  uint16(tunnelPort),
				LocalPort:   uint16(localPort),
				Connections: section.Int("connections", 1),
			}
			clients = append(clients, client)
			logger.Debugf("tunnel '%s' added", client.TunnelName)
		}
	}

	cl := len(clients)
	if cl == 0 {
		logger.Infof("no x.tunnel client config")
		return
	}

	for i, client := range clients {
		if i == cl-1 {
			client.Run()
		} else {
			go client.Run()
		}
	}
}
