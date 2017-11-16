package tunnel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const (
	httpPort        = 8088
	poxyHttpPort    = 8080
	tunnelPort      = 8087
	aesKey          = "xhka6uhfu2yvi7as"
	maxConncectines = 32
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
			Port:   tunnelPort,
			Secret: aesKey,
		}

		err := serv.AddTunnel("http-test", poxyHttpPort, maxConncectines)
		if err != nil {
			log.Error(err)
			return
		}

		err = serv.Serve()
		if err != nil {
			log.Error(err)
		}
	}()

	go func() {
		client := &Client{
			Server:       fmt.Sprintf("127.0.0.1:%d", tunnelPort),
			ServerSecret: aesKey,
			TunnelName:   "http-test",
			ForwardPort:  httpPort,
		}

		client.Run()
	}()
}

func Test(t *testing.T) {
	time.Sleep(time.Second)
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
	time.Sleep(3 * time.Second)
}
