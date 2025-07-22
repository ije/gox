package log

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"
)

func TestFileFS(t *testing.T) {
	logFileName := path.Join(os.TempDir(), fmt.Sprintf("gox-test-%s.log", time.Now().Format("2006-01-02-03")))
	os.Remove(logFileName)

	log, err := New("file:" + path.Join(os.TempDir(), "gox-test.log") + "?buffer=64&maxFileSize=2kb&fileDateFormat=2006-01-02-03&term")
	if err != nil {
		t.Fatal(err)
	}

	wr, ok := log.output.(*fileWriter)
	if !ok {
		t.Fatal("not a file log")
	}
	if log.bufcap != 64 {
		t.Fatalf("invalid buffer cap %d, should be %d", log.bufcap, 64)
	}
	if !log.term {
		t.Fatal("term should be enabled")
	}
	if wr.maxFileSize != 2*1024 {
		t.Fatalf("invalid maxFileSize %d, should be %d", wr.maxFileSize, 2*1024)
	}
	if wr.fileDateFormat != "2006-01-02-03" {
		t.Fatalf("invalid gfileDateFormat %s, should be %s", wr.fileDateFormat, "2006-01-02-03")
	}

	log.Print("Hello World!")
	log.Debug(":D")
	log.Info("Ok")
	log.Warn("BEEP")
	log.Error("BOOM!!!")

	exp := `2016/01/02 15:04:05 Hello World!
2016/01/02 15:04:05 [debug] :D
2016/01/02 15:04:05 [info] Ok
2016/01/02 15:04:05 [warn] BEEP
2016/01/02 15:04:05 [error] BOOM!!!
`

	data, err := os.ReadFile(logFileName)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(exp)-log.buflen {
		t.Fatalf("invalid file size %d, should be %d", len(data), len(exp)-log.buflen)
	}

	log.FlushBuffer()
	if log.buflen != 0 {
		t.Fatalf("invalid buffer len %d, should be %d", log.buflen, 0)
	}

	data, err = os.ReadFile(logFileName)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(exp) {
		t.Fatalf("invalid file size %d, should be %d", len(data), len(exp))
	}

	logText := "Dolore magna aliquam erat volutpat ut wisi enim ad minim veniam quis, nostrud exerci tation ullamcorper."
	log.Print(logText)
	if log.buflen != 0 {
		t.Fatalf("invalid buffer len %d, should be %d", log.buflen, 0)
	}

	data, err = os.ReadFile(logFileName)
	if err != nil {
		t.Fatal(err)
	}
	if l := len(exp) + len("2016/01/02 15:04:05 ") + len(logText) + 1; len(data) != l {
		t.Fatalf("invalid file size %d, should be %d", len(data), l)
	}
}
