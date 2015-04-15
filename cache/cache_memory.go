package cache

import (
	"runtime"
	"strconv"
	"sync"
	"time"
)

var mcPool = map[string]*mCache{}
var mcPoolLock sync.Mutex

func getMCache(region string) (mc *mCache) {
	mcPoolLock.Lock()
	defer mcPoolLock.Unlock()

	var ok bool
	mc, ok = mcPool[region]
	if !ok {
		mc = &mCache{region: region, values: map[string]mcValue{}}
		mcPool[region] = mc
	}
	return
}

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

func (mc *mCache) SetRegion(region string) error {
	if mc.region == region {
		return nil
	}

	_mc := getMCache(region)

	mc.lock.Lock()
	_mc.lock.Lock()
	gcInterval, _gcInterval := _mc.gcInterval, mc.gcInterval
	mc.region, _mc.region = region, mc.region
	mc.values, _mc.values = _mc.values, mc.values
	mc.lock.Unlock()
	_mc.lock.Unlock()

	mc.setGCInterval(gcInterval)
	_mc.setGCInterval(_gcInterval)

	mcPoolLock.Lock()
	mcPool[_mc.region], mcPool[mc.region] = _mc, mc
	mcPoolLock.Unlock()

	return nil
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

func (mc *mCache) Add(key string, val interface{}, lifetime time.Duration) error {
	mc.lock.RLock()
	_, ok := mc.values[key]
	mc.lock.RUnlock()

	if !ok {
		mc.Set(key, val, lifetime)
	}
	return nil
}

func (mc *mCache) Set(key string, val interface{}, lifetime time.Duration) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if lifetime > 0 {
		mc.values[key] = mcValue{time.Now().Add(lifetime).UnixNano(), val}
	} else {
		mc.values[key] = mcValue{0, val}
	}
	return nil
}

func (mc *mCache) Delete(key string) error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	delete(mc.values, key)
	return nil
}

func (mc *mCache) Flush() error {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	mc.values = map[string]mcValue{}
	return nil
}

func (mc *mCache) Has(key string) (ok bool, err error) {
	mc.lock.RLock()
	defer mc.lock.RUnlock()

	_, ok = mc.values[key]
	return
}

func (mc *mCache) Keys() (keys []string, err error) {
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

func (mc *mCache) setGCInterval(interval time.Duration) *mCache {
	mc.lock.Lock()
	defer mc.lock.Unlock()

	if interval != mc.gcInterval {
		if mc.gcTimer != nil {
			mc.gcTimer.Stop()
		}
		mc.gcInterval = interval
		if interval > 0 {
			mc.gcTimer = time.AfterFunc(interval, mc.gc)
		}
	}
	return mc
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

type mcDriver struct{}

func (mcd *mcDriver) Open(region string, args map[string]string) (cache Cache, err error) {
	var gcInterval int
	if s, ok := args["gcInterval"]; ok && len(s) > 0 {
		if gcInterval, err = strconv.Atoi(s); err != nil {
			return
		}
	}

	cache = getMCache(region).setGCInterval(time.Duration(gcInterval) * time.Second)
	return
}

func init() {
	Register("memory", &mcDriver{})
}
