package tunnel

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

const httpPort = 8088
const poxyHttpPort = 8080
const tunnelPort = 8087
const aesKey = "hello"

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
		client := &Client{
			Server:      fmt.Sprintf("127.0.0.1:%d", tunnelPort),
			AESKey:      aesKey,
			ServiceName: "http",
			ServicePort: httpPort,
			Connections: 8,
		}

		err := client.Listen()
		if err != nil {
			log.Error("client listen:", err)
		}
	}()

	go func() {
		serv := &Server{
			Port:   tunnelPort,
			AESKey: aesKey,
		}

		err := serv.AddService("http", poxyHttpPort, 8)
		if err != nil {
			log.Error(err)
			return
		}

		err = serv.Serve()
		if err != nil {
			log.Error(err)
		}
	}()
}

func Test(t *testing.T) {
	for i := 0; i < 1000; i++ {
		time.Sleep(5 * time.Millisecond)
		go func(i int) {
			r, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", poxyHttpPort))
			if err != nil {
				t.Log("http get:", err)
				return
			}

			ret, _ := ioutil.ReadAll(r.Body)
			if string(ret) != "Hello World" {
				t.Fatal(ret)
			}
		}(i)
	}

	time.Sleep(time.Second)
}
