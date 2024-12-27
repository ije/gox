package sync

import (
	"sync"
	"sync/atomic"
)

// KeyedMutex is a mutex that locks on a key.
type KeyedMutex struct {
	mutexes sync.Map
}

// KeyedMutexItem is an item in a KeyedMutex.
type KeyedMutexItem struct {
	lock  sync.Mutex
	count atomic.Int32
}

// Lock locks the mutex for the given key.
func (m *KeyedMutex) Lock(key string) func() {
	value, _ := m.mutexes.LoadOrStore(key, &KeyedMutexItem{})
	mtx := value.(*KeyedMutexItem)
	mtx.count.Add(1)
	mtx.lock.Lock()

	return func() {
		mtx.lock.Unlock()
		if mtx.count.Add(-1) == 0 {
			m.mutexes.Delete(key)
		}
	}
}

// Once is an object that will perform exactly one action.
// Different from sync.Once, this implementation allows the once function
// to return an error, that doesn't update the done flag.
type Once struct {
	done atomic.Uint32
	lock sync.Mutex
}

// Do calls the function f if and only if Do is being called for the
// first time for this instance of [Once]
func (o *Once) Do(f func() error) error {
	if o.done.Load() == 0 {
		// Outlined slow-path to allow inlining of the fast-path.
		return o.doSlow(f)
	}
	return nil
}

func (o *Once) doSlow(f func() error) error {
	o.lock.Lock()
	defer o.lock.Unlock()
	if o.done.Load() == 0 {
		err := f()
		if err == nil {
			o.done.Store(1)
		}
		return err
	}
	return nil
}
