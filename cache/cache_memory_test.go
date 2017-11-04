package cache

import (
	"testing"
	"time"
)

func TestMCache(t *testing.T) {
	cache, err := New("memory:region1?gcInterval=300")
	if err != nil {
		t.Error(err)
		return
	}

	cache.Set("key", "hello world", 0)
	t.Logf("%#v", cache)
	t.Log(cache.Get("key"))

	cache.Add("key", "hello world!", 0)
	t.Log(cache.Get("key"))

	cache.Set("key", "hello world!", 3*time.Second)
	t.Log(cache.Get("key"))

	time.Sleep(3 * time.Second)
	t.Log(cache.Get("key"))
	cache.Set("key", "hello world 2", 0)

	cache.SetRegion("region2")
	cache.Set("key", "hello world~", 0)
	t.Logf("%#v", cache)
	t.Log(cache.Get("key"))
	t.Logf("%#v", mcPool)

	cache.SetRegion("hello")
	t.Logf("%#v", cache)
}
