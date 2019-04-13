package main

import (
	"log"
	"strings"

	"github.com/ije/gox/debug"
)

func main() {
	err := debug.AddProcess(&debug.Process{
		Name: "godoc",
		Path: "godoc",
		Args: []string{"-http=:6066"},
		TermColorManager: func(b []byte) debug.TermColor {
			return debug.T_COLOR_RED
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = debug.UseHttpProxy(80, true, "127.0.0.1:6066", "http")
	if err != nil {
		log.Fatal(err)
	}

	debug.AddCommand("say", func(args ...string) (ret string, err error) {
		ret = "('" + strings.Join(args, "', '") + "')"
		return
	})

	debug.Run()
}
