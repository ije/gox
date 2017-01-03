package tunnel

import (
	"net/http"
	"testing"
)

var aesKey = "hello"

func Test(t *testing.T) {
	go http.ListenAndServe(":8088", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("server", "go.test")
		w.Write([]byte("Hello World!"))
	}))

	serv := &Server{
		Port:   8090,
		AESKey: aesKey,
	}

	err := serv.AddService("http", 8888)
	if err != nil {
		log.Error(err)
		return
	}

	go func(serv *Server) {
		err := serv.Serve()
		if err != nil {
			log.Error(err)
		}
	}(serv)

	client := &Client{
		Server:      ":8090",
		AESKey:      aesKey,
		ServiceName: "http",
		ServicePort: 8088,
	}
	err = client.Listen()
	if err != nil {
		log.Error("client listen:", err)
	}
}
