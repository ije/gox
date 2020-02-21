package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
	"github.com/ije/gox/utils"
)

type Config struct {
	Server  string   `json:"server"`
	Tunnels []Tunnel `json:"tunnels"`
}

type Tunnel struct {
	Server           string `json:"server"`
	Name             string `json:"name"`
	Port             uint16 `json:"port"`
	ForwardPort      uint16 `json:"forwardPort"`
	MaxProxyLifetime int    `json:"maxProxyLifetime"`
}

func main() {
	cfile := flag.String("c", "/etc/gox.tunnel/config.json", "gox tunnel client configuration")
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

	var tunnelCount int
	for _, t := range config.Tunnels {
		if len(t.Name) > 0 && t.ForwardPort > 0 && t.Port > 0 {
			server := config.Server
			if t.Server != "" {
				server = t.Server
			}
			server = strings.TrimSpace(server)
			if server != "" {
				tc := &tunnel.Client{
					Server: server,
					Tunnel: tunnel.Tunnel{
						Name:             t.Name,
						Port:             t.Port,
						MaxProxyLifetime: t.MaxProxyLifetime,
					},
					ForwardPort: t.ForwardPort,
				}
				go tc.Connect()
				tunnelCount++
			}
		}
	}

	if tunnelCount > 0 {
		logger.Infof("%d tunnels added", tunnelCount)
		utils.WaitExit(func(sig os.Signal) bool {
			return true
		})
	} else {
		logger.Error("exit: no tunnels")
	}
}
