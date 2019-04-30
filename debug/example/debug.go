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
		TermLineColor: func(p []byte) debug.TermColor {
			return debug.T_COLOR_PURPLE
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = debug.UseHttpProxy(80, "127.0.0.1:6066", true)
	if err != nil {
		log.Fatal(err)
	}

	debug.AddCommand("say", func(args ...string) (ret string, err error) {
		ret = "('" + strings.Join(args, "', '") + "')"
		return
	})

	debug.Run()
}
