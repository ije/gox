package tunnel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	tunnelPort    = 8087
	httpPort      = 8088
	httpProxyPort = 8089
)

func init() {
	heartBeatInterval = 1

	// start a http server
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", httpPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World!"))
		}),
	}
	s.SetKeepAlivesEnabled(false)
	go s.ListenAndServe()

	// tunnel server
	serv := &Server{
		Port: tunnelPort,
	}
	go serv.Serve()

	// tunnel client
	client := &Client{
		Server: fmt.Sprintf("127.0.0.1:%d", tunnelPort),
		Tunnel: Tunnel{
			Name: "test",
			Port: httpProxyPort,
		},
		ForwardPort: httpPort,
	}
	go client.Connect()
}

func Test(t *testing.T) {
	time.Sleep(3 * time.Second) // wait init finished

	for i := 0; i < 100; i++ {
		go func() {
			r, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", httpProxyPort))
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
	time.Sleep(3 * time.Second)
}
