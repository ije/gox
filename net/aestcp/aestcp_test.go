package aestcp

import (
	"fmt"
	"testing"
	"time"
)

var key = []byte("example key 0123456789")

func listen() {
	ln, err := Listen("tcp", ":8080", key)
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		_, err = conn.Write([]byte("Hello from server!"))
		if err != nil {
			panic(err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		fmt.Println("server:", string(buf[:n]))
	}
}

func dial() {
	conn, err := Dial("tcp", "127.0.0.1:8080", key)
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println("client:", string(buf[:n]))

	_, err = conn.Write([]byte("Hello from client!"))
	if err != nil {
		panic(err)
	}
}

func Test(t *testing.T) {
	go listen()
	time.Sleep(time.Second * 1)
	dial()
	time.Sleep(time.Second * 1)
}
