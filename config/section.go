package config

import (
	"fmt"
	"strconv"
	"strings"

	strconv2 "github.com/ije/gox/strconv"
)

type Section map[string]string

func (section Section) IsEmpty() bool {
	return len(section) == 0
}

func (section Section) Contains(key string) (ok bool) {
	_, ok = section[key]
	return
}

func (section Section) String(key string, def string) string {
	if val, ok := section[key]; ok {
		return val
	}
	return def
}

func (section Section) Bool(key string, def bool) bool {
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

func (section Section) Int(key string, def int) int {
	if val, ok := section[key]; ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
		delete(section, key)
	}
	return def
}

func (section Section) Int64(key string, def int64) int64 {
	if val, ok := section[key]; ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
		delete(section, key)
	}
	return def
}

func (section Section) Bytes(key string, def int64) int64 {
	if val, ok := section[key]; ok {
		if i, err := strconv2.ParseByte(val); err == nil {
			return i
		}
		delete(section, key)
	}
	return def
}

func (section Section) Float64(key string, def float64) float64 {
	if val, ok := section[key]; ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
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
