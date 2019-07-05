package cache

import (
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/ije/gox/utils"
)

type mData struct {
	data     interface{}
	deadline int64
}

func (v mData) isExpired() bool {
	return v.deadline > 0 && time.Now().UnixNano() > v.deadline
}

type mCache struct {
	lock       sync.RWMutex
	storage    map[string]mData
	gcInterval time.Duration
	gcTimer    *time.Timer
}

func (mc *mCache) Has(key string) (ok bool, err error) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()

	_, ok = mc.storage[key]
	return
}

func (mc *mCache) Get(key string) (data interface{}, err error) {
	mc.lock.RLock()
	s, ok := mc.storage[key]
	mc.lock.RUnlock()
	if !ok {
		err = ErrNotFound
		return
	}

	if s.isExpired() {
		mc.Delete(key)
		err = ErrExpired
		return
	}

	data = s.data
	return
}

func (mc *mCache) Set(key string, data interface{}) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.storage[key] = mData{data, 0}
	return nil
}

func (mc *mCache) SetTemp(key string, data interface{}, lifetime time.Duration) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if lifetime > 0 {
		mc.storage[key] = mData{data, time.Now().Add(lifetime).UnixNano()}
	}
	return nil
}

func (mc *mCache) Delete(key string) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	delete(mc.storage, key)
	return nil
}

func (mc *mCache) Flush() error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.storage = map[string]mData{}
	return nil
}

func (mc *mCache) Run(name string, data ...interface{}) error {
	return nil
}

func (mc *mCache) setGCInterval(interval time.Duration) {
	if interval > 0 && interval != mc.gcInterval {
		if mc.gcTimer != nil {
			mc.gcTimer.Stop()
		}
		mc.gcInterval = interval
		if interval > time.Second {
			mc.gcTimer = time.AfterFunc(interval, mc.gc)
		}
	}
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
	gcInterval := 30 * time.Minute
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
