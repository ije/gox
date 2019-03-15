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
	poxyHttpPort = 8089
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

	serv := &Server{
		Port: tunnelPort,
	}
	go serv.Serve()

	client := &Client{
		Server: fmt.Sprintf("127.0.0.1:%d", tunnelPort),
		Tunnel: Tunnel{
			Name:           "http-proxy-testing",
			Port:           poxyHttpPort,
			MaxConnections: 100,
		},
		ForwardPort: httpPort,
	}
	go client.Connect()
}

func Test(t *testing.T) {
	time.Sleep(time.Second) // wait init

	for i := 0; i < 100; i++ {
		go func() {
			r, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", poxyHttpPort))
			if err != nil {
				t.Fatal(err)
			}
			defer r.Body.Close()

			ret, _ := ioutil.ReadAll(r.Body)
			if string(ret) != "Hello World!" {
				t.Fatal(string(ret))
			}
		}()
	}
	time.Sleep(15 * time.Second)
}
