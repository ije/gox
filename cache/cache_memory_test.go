package cache

import (
	"testing"
	"time"
)

func TestMCache(t *testing.T) {
	cache, err := New("memory:?gcInterval=300")
	if err != nil {
		t.Error(err)
		return
	}

	cache.Put(nil, "key", []byte("hello world"))
	data, err := cache.Get(nil, "key")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Fatal("bad data", string(data))
	}

	cache.PutTemp(nil, "keytemp", []byte("hello world"), 3*time.Second)
	_, err = cache.Get(nil, "keytemp")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)
	_, err = cache.Get(nil, "keytemp")
	if err != ErrExpired {
		t.Fatal("should be expired error, but", err)
	}
}
