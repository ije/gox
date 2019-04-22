package log

import (
	"testing"
)

func TestFileLogger(t *testing.T) {
	log, err := New("file:/tmp/test.log?maxBytes=2kb&fileDateFormat=2006-01-02-15")
	if err != nil {
		t.Fatal(err)
	}

	wr, ok := log.output.(*fileWriter)
	if !ok {
		t.Fatal("not a memory cache")
	}
	if wr.maxFileBytes != 2*1024 {
		t.Fatalf("invalid gc interval %d, should be %d", wr.maxFileBytes, 2*1024)
	}
	if wr.fileDateFormat != "2006-01-02-15" {
		t.Fatalf("invalid gc interval %s, should be %s", wr.fileDateFormat, "2006-01-02-15")
	}

	log.Print("Hello!")
	log.Log("info", "Hello!")
	log.Debug(":D")
	log.Info("Ok")
	log.Warn("No good")
	log.Error("ERROR")
	log.Fatal("FATAL")
	log.Info("Hello?")
}
