package cache

import (
	"sort"
)

var drivers = map[string]CacheDriver{}

type CacheDriver interface {
	Open(region string, args map[string]string) (cache Cache, err error)
}

func Register(name string, driver CacheDriver) {
	if driver == nil {
		panic("cache: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("cache: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func Drivers() (list []string) {
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return
}
