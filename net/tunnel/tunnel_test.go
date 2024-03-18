package tunnel

import (
	"fmt"
	"io"
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
			w.Write([]byte("Hello world!"))
		}),
	}
	s.SetKeepAlivesEnabled(false)
	go s.ListenAndServe()

	// tunnel server
	serv := &Server{
		Port:     tunnelPort,
		Password: "1234",
	}
	go serv.Serve()

	// tunnel client
	client := &Client{
		Server:   fmt.Sprintf("127.0.0.1:%d", tunnelPort),
		Password: "1234",
		Tunnel: &TunnelProps{
			Name: "test-tunnel",
			Port: httpProxyPort,
		},
		ForwardPort: httpPort,
	}
	go client.Connect()
}

func Test(t *testing.T) {
	time.Sleep(time.Second / 10) // wait for server and client to start

	for i := 0; i < 100; i++ {
		r, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", httpProxyPort))
		if err != nil {
			t.Fatal(err)
		}
		defer r.Body.Close()

		ret, _ := io.ReadAll(r.Body)
		if string(ret) != "Hello world!" {
			t.Fatal(string(ret))
		}
	}
}
