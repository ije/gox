package log

import (
	"testing"
)

func TestFileLogger(t *testing.T) {
	log, err := New("file:/tmp/go.test.log?maxBytes=2kb")
	if err != nil {
		t.Fatal(err)
	}
	log.Print("Hello World !")
	log.Debug("Hello World !")
	log.Info("Hello World !")
	log.Warn("Hello World !")
	log.Error("Hello World !")
	log.Fatal("\"Neque porro quisquam est qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit...\"")
	log.Info("Hello World !")
}
