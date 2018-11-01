package main

import (
	"flag"

	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
)

func main() {
	port := flag.Int("port", 333, "tunnel service port")
	httpPort := flag.Int("http-port", 8080, "tunnel http port")
	debug := flag.Bool("d", false, "debug mode")
	flag.Parse()

	logger := &log.Logger{}
	if !*debug {
		logger.SetLevelByName("info")
	}
	tunnel.SetLogger(logger)

	ts := &tunnel.Server{
		Port:     uint16(*port),
		HTTPPort: uint16(*httpPort),
	}
	ts.Serve()
}
