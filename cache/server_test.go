package cache

import (
	"testing"
)

func TestParseAddr(t *testing.T) {
	t.Log(splitNA("tcp6@127.0.0.1:8080"))
	t.Log(splitNA("unixpacket@127.0.0.1:8080"))
}
