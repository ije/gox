package cache

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Cache interface {
	Has(ctx context.Context, key string) (bool, error)
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, data []byte) error
	PutTemp(ctx context.Context, key string, data []byte, lifetime time.Duration) error
	Delete(ctx context.Context, key string) error
	Flush(ctx context.Context) error
	Run(ctx context.Context, name string, args ...[]byte) error
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
