package cache

import (
	"testing"
	"time"
)

func TestMCache(t *testing.T) {
	cache, err := New("memory?gcInterval=3s")
	if err != nil {
		t.Error(err)
		return
	}

	mc, ok := cache.(*mCache)
	if !ok {
		t.Fatal("not a memory cache")
	}
	if mc.gcInterval != 3*time.Second {
		t.Fatalf("invalid gc interval %v, should be %v", mc.gcInterval, 3*time.Second)
	}

	cache.Set("key", "hello world")
	value, err := cache.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if value != "hello world" {
		t.Fatalf("invalid value(%v), shoud be 'hello world'", value)
	}

	cache.Set("keytemp", []byte("hello world"), 3*time.Second)
	_, err = cache.Get("keytemp")
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)
	_, err = cache.Get("keytemp")
	if err != ErrExpired {
		t.Fatal("should be expired error, but", err)
	}
}
