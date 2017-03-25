package tunnel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	serverPort   = 8087
	tunnelPort   = 8080
	httpPort     = 8088
	aesKey       = "hello"
	conncectines = 16
)

func init() {
	log.SetLevelByName("debug")

	s := &http.Server{
		Addr: fmt.Sprintf(":%d", httpPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World"))
		}),
	}
	s.SetKeepAlivesEnabled(false)
	go s.ListenAndServe()

	go func() {
		serv := &Server{
			Port:   serverPort,
			AESKey: aesKey,
		}

		err := serv.Serve()
		if err != nil {
			log.Error(err)
		}
	}()

	go func() {
		client := &Client{
			Server:      fmt.Sprintf("127.0.0.1:%d", serverPort),
			AESKey:      aesKey,
			TunnelName:  "http",
			TunnelPort:  tunnelPort,
			LocalPort:   httpPort,
			Connections: conncectines,
		}

		client.Run()
	}()
}

func Test(t *testing.T) {
	time.Sleep(time.Second / 10)

	for i := 0; i < 1000; i++ {
		go func() {
			r, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", tunnelPort))
			if err != nil {
				t.Fatal(err)
			}

			ret, _ := ioutil.ReadAll(r.Body)
			if string(ret) != "Hello World" {
				t.Fatal(string(ret))
			}
		}()
		time.Sleep(time.Millisecond)
	}
}
