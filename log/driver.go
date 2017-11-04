package log

import (
	"io"
	"sort"
)

var drivers = map[string]Driver{}

type Driver interface {
	Open(addr string, args map[string]string) (writer io.Writer, err error)
}

func Register(name string, driver Driver) {
	if driver == nil {
		panic("log: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("log: Register called twice for driver " + name)
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
