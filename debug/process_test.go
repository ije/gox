package debug

import (
	"testing"
	"time"
)

func TestProcess(t *testing.T) {
	err := AddHttpProxyProcess(":80", map[string]string{"godoc": "127.0.0.1:6060"})
	if err != nil {
		t.Fatal(err)
	}

	var proc *Process
	for _, process := range processes {
		if process.Name == "gox.debug.http-proxy" {
			proc = process
			break
		}
	}
	if proc == nil {
		t.Fatal()
	}

	err = proc.Build()
	if err != nil {
		t.Fatal("build: ", err)
	}

	err = proc.Start()
	if err != nil {
		t.Fatal("start: ", err)
	}

	time.Sleep(9 * time.Second)

	err = proc.Stop()
	if err != nil {
		t.Fatal("stop: ", err)
	}
}
