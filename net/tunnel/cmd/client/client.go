package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
	"github.com/ije/gox/utils"
)

type Config struct {
	Server  string   `json:"server"`
	Clients []Client `json:"clients"`
}

type Client struct {
	Name             string `json:"name"`
	Port             uint16 `json:"port"`
	ForwardPort      uint16 `json:"forwardPort"`
	MaxConnections   int    `json:"maxConnections"`
	MaxProxyLifetime int    `json:"maxProxyLifetime"`
}

func main() {
	cfile := flag.String("c", "/etc/tunnel/client.json", "tunnel client configuration file")
	debug := flag.Bool("d", false, "debug mode")
	flag.Parse()

	var config Config
	err := utils.ParseJSONFile(*cfile, &config)
	if err != nil && os.IsExist(err) {
		fmt.Println("load the configuration failed:", err)
		return
	}

	logger := &log.Logger{}
	if !*debug {
		logger.SetLevelByName("info")
	}
	tunnel.SetLogger(logger)

	clients := len(config.Clients)
	for _, client := range config.Clients {
		if len(client.Name) > 0 && client.ForwardPort > 0 && client.Port > 0 {
			tc := &tunnel.Client{
				Server: config.Server,
				Tunnel: tunnel.Tunnel{
					Name:             client.Name,
					Port:             client.Port,
					MaxConnections:   client.MaxConnections,
					MaxProxyLifetime: client.MaxProxyLifetime,
				},
				ForwardPort: client.ForwardPort,
			}
			go tc.Connect()
		}
	}

	if clients > 0 {
		utils.WaitExit(func(sig os.Signal) bool {
			return true
		})
	}
}
