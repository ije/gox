package tunnel

import (
	"testing"
)

var aesKey = "hello"

func Test(t *testing.T) {
	serv := &Server{
		Port:   8080,
		AESKey: aesKey,
	}

	err := serv.AddService("ssh", 2222)
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
		Server:      ":8080",
		AESKey:      aesKey,
		ServiceName: "ssh",
		ServicePort: 22,
	}
	err = client.Listen()
	if err != nil {
		log.Error("client listen:", err)
	}
}
