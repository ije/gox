package cache

import (
	"testing"
)

func TestClient(t *testing.T) {
	cs := make(chan bool, 1)
	cs <- true
	t.Log(len(cs), cap(cs))
	<-cs
	t.Log(len(cs), cap(cs))
}
