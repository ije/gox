package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/ije/gox/net/tunnel"
)

func main() {
	port := flag.Int("port", 333, "tunnel service port")
	password := flag.String("password", "", "tunnel service password")
	httpPort := flag.Int("http-port", 8080, "tunnel service http server addr")
	flag.Parse()

	ts := &tunnel.Server{
		Port:     uint16(*port),
		Password: *password,
	}
	go http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), ts)
	ts.Serve()
}
