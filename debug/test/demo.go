package main

import (
	"log"

	"github.com/ije/gox/debug"
	"github.com/ije/gox/term"
)

func main() {
	err := debug.AddProcess(&debug.Process{
		Name: "godoc",
		Path: "godoc",
		Args: []string{"-http=:6066"},
		TermColorManager: func(b []byte) term.Color {
			return term.COLOR_RED
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = debug.UseHttpProxy(map[string]string{
		"godoc":    "127.0.0.1:6066",
		"s.g.com":  "127.0.0.1:8080",
		".g.com":   "127.0.0.1:8080",
		".g-*.com": "127.0.0.1:8080",
		"*.g.com":  "127.0.0.1:8080",
	})
	if err != nil {
		log.Fatal(err)
	}

	debug.AddCommand("test", func(args ...string) (ret string, err error) {
		log.Println("hello world", args)
		return
	})

	debug.Run()
}
