package cache

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrExpired  = errors.New("expired")
)

type Cache interface {
	SetRegion(region string) error
	Get(key string) (v interface{}, err error)
	Has(key string) (ok bool, err error)
	Add(key string, val interface{}, lifetime time.Duration) error
	Set(key string, val interface{}, lifetime time.Duration) error
	Keys() (keys []string, err error)
	Delete(key string) error
	Flush() error
}

func New(url string) (cache Cache, err error) {
	i := strings.IndexByte(url, ':')
	if i == -1 {
		err = errors.New("Incorrect URL '" + url + "'")
		return
	}
	driver, ok := drivers[strings.ToLower(url[:i])]
	if !ok {
		err = errors.New("Unknown driver '" + url[:i] + "'")
		return
	}
	url = url[i+1:]

	addr := url
	args := map[string]string{}
	if i = strings.IndexByte(url, '?'); i >= 0 {
		for _, q := range strings.Split(url[i+1:], "&") {
			kv := strings.SplitN(q, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := ""
			if len(kv) == 2 {
				value = strings.TrimSpace(kv[1])
			}
			args[key] = value
		}
		addr = url[:i]
	}

	cache, err = driver.Open(addr, args)
	return
}
