package main

import (
	"log"

	"github.com/ije/gox/debug"
)

func main() {
	err := debug.AddProcess(&debug.Process{
		Name: "godoc",
		Path: "godoc",
		Args: []string{"-http", ":6060"},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = debug.AddHttpProxyProcess(":80", map[string]string{"godoc": "127.0.0.1:6060"})
	if err != nil {
		log.Fatal(err)
	}

	debug.AddCommand("cmd", func(args ...string) (ret string, err error) {
		log.Println("hello world", args)
		return
	})

	debug.Run()
}
