package tunnel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	tunnelPort   = 8087
	httpPort     = 8088
	poxyHttpPort = 8080
)

func init() {
	log.SetLevelByName("debug")

	s := &http.Server{
		Addr: fmt.Sprintf(":%d", httpPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World!"))
		}),
	}
	s.SetKeepAlivesEnabled(false)
	go s.ListenAndServe()

	go func() {
		serv := &Server{
			Port: tunnelPort,
		}

		err := serv.Serve()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		client := &Client{
			Server: fmt.Sprintf("127.0.0.1:%d", tunnelPort),
			Tunnel: Tunnel{
				Name:           "http-proxy-testing",
				Port:           poxyHttpPort,
				MaxConnections: 100,
			},
			ForwardPort: httpPort,
		}
		client.Connect()
	}()
}

func Test(t *testing.T) {
	time.Sleep(time.Second / 2) // wait init goruntines end

	for i := 0; i < 1000; i++ {
		r, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", poxyHttpPort))
		if err != nil {
			t.Fatal(err)
		}

		ret, _ := ioutil.ReadAll(r.Body)
		if string(ret) != "Hello World!" {
			t.Fatal(string(ret))
		}
	}
	time.Sleep(12 * time.Second)
}
