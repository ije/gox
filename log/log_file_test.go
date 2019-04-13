package log

import (
	"testing"
)

func TestFileLogger(t *testing.T) {
	log, err := New("file:/tmp/test.log?maxBytes=2kb&fileDate")
	if err != nil {
		t.Fatal(err)
	}
	log.Print("Hello!")
	log.Log("info", "Hello!")
	log.Debug(":D")
	log.Info("OK")
	log.Warn("NOT GOOD")
	log.Error("ERROR")
	log.Fatal("DIE")
	log.Info("Hello?")
}
