package config

import (
	"fmt"
	"strings"

	"github.com/ije/gox/strconv"
)

type Section map[string]string

func (section Section) IsEmpty() bool {
	return len(section) == 0
}

func (section Section) Contains(key string) (ok bool) {
	_, ok = section[key]
	return
}

func (section Section) Keys() (keys []string) {
	var i int
	keys = make([]string, len(section))
	for key, _ := range section {
		keys[i] = key
		i++
	}
	return
}

func (section Section) Each(hanlder func(key string, value string)) {
	for key, value := range section {
		hanlder(key, value)
	}
}

func (section Section) String(key string, extra ...string) string {
	var def string
	if len(extra) > 0 {
		def = extra[0]
	}
	if val, ok := section[key]; ok {
		return val
	}
	return def
}

func (section Section) Bool(key string, extra ...bool) bool {
	var def bool
	if len(extra) > 0 {
		def = extra[0]
	}
	if val, ok := section[key]; ok {
		switch strings.ToLower(val) {
		case "false", "0", "no", "off", "disable":
			return false
		case "true", "1", "yes", "on", "enable":
			return true
		}
	}
	return def
}

func (section Section) Int(key string, extra ...int) int {
	var def int
	if len(extra) > 0 {
		def = extra[0]
	}
	if val, ok := section[key]; ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		delete(section, key)
	}
	return def
}

func (section Section) Int64(key string, extra ...int64) int64 {
	var def int64
	if len(extra) > 0 {
		def = extra[0]
	}
	if val, ok := section[key]; ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
		delete(section, key)
	}
	return def
}

func (section Section) Float64(key string, extra ...float64) float64 {
	var def float64
	if len(extra) > 0 {
		def = extra[0]
	}
	if val, ok := section[key]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
		delete(section, key)
	}
	return def
}

func (section Section) Bytes(key string, extra ...int64) int64 {
	var def int64
	if len(extra) > 0 {
		def = extra[0]
	}
	if val, ok := section[key]; ok {
		if i, err := strconv.ParseBytes(val); err == nil {
			return i
		}
		delete(section, key)
	}
	return def
}

func (section Section) Set(key string, value interface{}) {
	switch v := value.(type) {
	case string:
		section[key] = v
	case int:
		section[key] = strconv.Itoa(v)
	case int64:
		section[key] = strconv.FormatInt(v, 10)
	case bool:
		if v {
			section[key] = "true"
		} else {
			section[key] = "false"
		}
	default:
		section[key] = fmt.Sprintf("%v", value)
	}
}
