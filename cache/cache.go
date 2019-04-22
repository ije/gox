package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ije/gox/utils"
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

// New returns a new cache by url
func New(url string) (cache Cache, err error) {
	path, query := utils.SplitByFirstByte(url, '?')
	name, addr := utils.SplitByFirstByte(path, ':')
	driver, ok := drivers[strings.ToLower(name)]
	if !ok {
		err = fmt.Errorf("Unknown driver '%s'", name)
		return
	}

	args := map[string]string{}
	for _, q := range strings.Split(query, "&") {
		k, v := utils.SplitByFirstByte(q, '=')
		if len(k) > 0 {
			args[k] = v
		}
	}

	cache, err = driver.Open(addr, args)
	return
}
