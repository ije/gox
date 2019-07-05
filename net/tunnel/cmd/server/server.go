package main

import (
	"flag"
	"net/http"

	"github.com/ije/gox/log"
	"github.com/ije/gox/net/tunnel"
)

func main() {
	port := flag.Int("port", 333, "tunnel service port")
	httpAddr := flag.String("http", "localhost:8080", "tunnel http server addr")
	debug := flag.Bool("d", false, "debug mode")
	flag.Parse()

	logger := &log.Logger{}
	if !*debug {
		logger.SetLevelByName("info")
	}
	tunnel.SetLogger(logger)

	ts := &tunnel.Server{
		Port: uint16(*port),
	}
	go http.ListenAndServe(*httpAddr, ts)
	for {
		ts.Serve()
	}
}
