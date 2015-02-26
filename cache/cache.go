package cache

import (
	"errors"
	"runtime"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExpired  = errors.New("expired")
)

type Cache interface {
	SetRegion(region string)
	Get(key string) (v interface{}, err error)
	Add(key string, val interface{}, lifetime time.Duration)
	Set(key string, val interface{}, lifetime time.Duration)
	Delete(key string)
	Flush()
	Has(key string) (ok bool)
	Keys() (keys []string)
	SetGCInterval(interval time.Duration)
}

var mcPool = map[string]*mCache{}
var mcPoolLock sync.Mutex

type mcValue struct {
	expired int64
	value   interface{}
}

func (val mcValue) Value() interface{} {
	if val.expired > 0 && time.Now().UnixNano() >= val.expired {
		return nil
	}
	return val.value
}

type mCache struct {
	lock       sync.RWMutex
	region     string
	values     map[string]mcValue
	gcTimer    *time.Timer
	gcInterval time.Duration
}

func (mc *mCache) SetRegion(region string) {
	_mc := New(region, 0).(*mCache)

	mc.lock.Lock()
	_mc.lock.Lock()
	mc.region, _mc.region = _mc.region, mc.region
	mc.values, _mc.values = _mc.values, mc.values
	gcInterval, _gcInterval := _mc.gcInterval, mc.gcInterval
	mc.gcInterval, _mc.gcInterval = 0, 0
	mc.lock.Unlock()
	_mc.lock.Unlock()
	mc.SetGCInterval(gcInterval)
	_mc.SetGCInterval(_gcInterval)

	mcPoolLock.Lock()
	mcPool[mc.region], mcPool[region] = _mc, mc
	mcPoolLock.Unlock()
}

func (mc *mCache) Get(key string) (v interface{}, err error) {
	mc.lock.RLock()
	val, ok := mc.values[key]
	mc.lock.RUnlock()
	if !ok {
		err = ErrNotFound
	} else if v = val.Value(); v == nil {
		mc.Delete(key)
		err = ErrExpired
	}
	return
}

func (mc *mCache) Add(key string, val interface{}, lifetime time.Duration) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if _, ok := mc.values[key]; !ok {
		mc.values[key] = mcValue{time.Now().Add(lifetime).UnixNano(), val}
	}
}

func (mc *mCache) Set(key string, val interface{}, lifetime time.Duration) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.values[key] = mcValue{time.Now().Add(lifetime).UnixNano(), val}
}

func (mc *mCache) Delete(key string) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	delete(mc.values, key)
}

func (mc *mCache) Flush() {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.values = map[string]mcValue{}
}

func (mc *mCache) Has(key string) (ok bool) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()

	_, ok = mc.values[key]
	return
}

func (mc *mCache) Keys() (keys []string) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()

	var i int
	keys = make([]string, len(mc.values))
	for key, _ := range mc.values {
		keys[i] = key
		i++
	}
	return
}

func (mc *mCache) SetGCInterval(interval time.Duration) {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if interval != mc.gcInterval {
		if mc.gcTimer != nil {
			mc.gcTimer.Stop()
		}
		mc.gcInterval = interval
		if interval >= time.Second {
			mc.gcTimer = time.AfterFunc(interval, mc.gc)
		}
	}
}

func (mc *mCache) gc() {
	mc.gcTimer = time.AfterFunc(mc.gcInterval, mc.gc)

	mc.lock.RLock()
	defer mc.lock.RUnlock()
	for key, value := range mc.values {
		if time.Now().UnixNano() >= value.expired {
			mc.lock.RUnlock()
			mc.lock.Lock()
			delete(mc.values, key)
			mc.lock.Unlock()
			mc.lock.RLock()
		}
	}

	runtime.GC()
}

func New(region string, gcInterval time.Duration) Cache {
	mcPoolLock.Lock()
	defer mcPoolLock.Unlock()

	mc, ok := mcPool[region]
	if !ok {
		mc = &mCache{region: region, values: map[string]mcValue{}}
		mcPool[region] = mc
	}
	mc.SetGCInterval(gcInterval)
	return mc
}
