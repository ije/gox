package cache

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/ije/gox/utils"
)

type mData struct {
	deadtime int64
	data     []byte
}

func (v mData) isExpired() bool {
	return v.deadtime > 0 && time.Now().UnixNano() > v.deadtime
}

func (v mData) Bytes() []byte {
	if v.isExpired() {
		return nil
	}
	return v.data
}

type mCache struct {
	lock       sync.RWMutex
	storage    map[string]mData
	gcTimer    *time.Timer
	gcInterval time.Duration
}

func (mc *mCache) Has(ctx context.Context, key string) (ok bool, err error) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()

	_, ok = mc.storage[key]
	return
}

func (mc *mCache) Get(ctx context.Context, key string) (data []byte, err error) {
	mc.lock.RLock()
	s, ok := mc.storage[key]
	mc.lock.RUnlock()
	if !ok {
		err = ErrNotFound
		return
	}

	if s.isExpired() {
		mc.Delete(ctx, key)
		err = ErrExpired
		return
	}

	data = s.Bytes()
	return
}

func (mc *mCache) Set(ctx context.Context, key string, data []byte) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.storage[key] = mData{0, data}
	return nil
}

func (mc *mCache) SetTemp(ctx context.Context, key string, data []byte, lifetime time.Duration) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if lifetime > 0 {
		mc.storage[key] = mData{time.Now().Add(lifetime).UnixNano(), data}
	}
	return nil
}

func (mc *mCache) Delete(ctx context.Context, key string) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	delete(mc.storage, key)
	return nil
}

func (mc *mCache) Flush(ctx context.Context) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.storage = map[string]mData{}
	return nil
}

func (mc *mCache) Run(ctx context.Context, name string, data ...[]byte) error {
	return nil
}

func (mc *mCache) setGCInterval(interval time.Duration) *mCache {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if interval > 0 && interval != mc.gcInterval {
		if mc.gcTimer != nil {
			mc.gcTimer.Stop()
		}
		mc.gcInterval = interval
		if interval > time.Second {
			mc.gcTimer = time.AfterFunc(interval, mc.gc)
		}
	}
	return mc
}

func (mc *mCache) gc() {
	mc.gcTimer = time.AfterFunc(mc.gcInterval, mc.gc)

	mc.lock.RLock()
	defer mc.lock.RUnlock()
	for key, d := range mc.storage {
		if d.isExpired() {
			mc.lock.RUnlock()
			mc.lock.Lock()
			delete(mc.storage, key)
			mc.lock.Unlock()
			mc.lock.RLock()
		}
	}

	runtime.GC()
}

type mcDriver struct{}

func (mcd *mcDriver) Open(region string, args map[string]string) (cache Cache, err error) {
	var gcInterval time.Duration
	if s, ok := args["gcInterval"]; ok && len(s) > 0 {
		gcInterval, err = utils.ParseDuration(s)
		if err != nil {
			err = errors.New("Invalid GC interval")
			return
		}
	}

	c := &mCache{
		storage:    map[string]mData{},
		gcInterval: gcInterval,
	}
	c.setGCInterval(gcInterval)
	return c, nil
}

func init() {
	Register("memory", &mcDriver{})
}
