package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ije/gox/net/tunnel"
	"github.com/ije/gox/utils"
)

type Config struct {
	Server   string   `json:"server"`
	Password string   `json:"password"`
	Tunnels  []Tunnel `json:"tunnels"`
}

type Tunnel struct {
	Server           string `json:"server"`
	Password         string `json:"password"`
	Name             string `json:"name"`
	Port             uint16 `json:"port"`
	ForwardPort      uint16 `json:"forwardPort"`
	MaxProxyLifetime int    `json:"maxProxyLifetime"`
}

func main() {
	cfile := flag.String("c", "./config.json", "gox tunnel client configuration")
	flag.Parse()

	var config Config
	err := utils.ParseJSONFile(*cfile, &config)
	if err != nil && os.IsExist(err) {
		fmt.Println("load the configuration failed:", err)
		return
	}

	var tunnelCount int
	for _, t := range config.Tunnels {
		if len(t.Name) > 0 && len(t.Name) < 256 && t.ForwardPort > 0 && t.Port > 0 {
			server := config.Server
			password := config.Password
			if t.Server != "" {
				server = t.Server
				password = t.Password
			}
			if server == "" {
				fmt.Printf("invalid tunnel(%s) config: missing server\n", t.Name)
				continue
			}
			server = strings.TrimSpace(server)
			if server != "" {
				tc := &tunnel.Client{
					Server:   server,
					Password: password,
					Tunnel: &tunnel.TunnelProps{
						Name:             t.Name,
						Port:             t.Port,
						MaxProxyLifetime: uint32(t.MaxProxyLifetime),
					},
					ForwardPort: t.ForwardPort,
				}
				go tc.Connect()
				tunnelCount++
			}
		} else {
			fmt.Println("invalid tunnel config:", t)
		}
	}

	if tunnelCount > 0 {
		utils.WaitForExitSignal(func(sig os.Signal) bool {
			return true
		})
	} else {
		fmt.Println("exit: no tunnels")
	}
}
