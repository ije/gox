package config

import (
	"strconv"
	"strings"

	"github.com/ije/go/utils"
)

type Section map[string]string

func (section Section) IsEmpty() bool {
	return len(section) == 0
}

func (section Section) Contains(key string) bool {
	_, ok := section[key]
	return ok
}

func (section Section) String(key, def string) string {
	if val, ok := section[key]; ok {
		return val
	}
	return def
}

func (section Section) Int(key string, def int) int {
	if val, ok := section[key]; ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return def
}

func (section Section) Int64(key string, def int64) int64 {
	if val, ok := section[key]; ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return def
}

func (section Section) Bytes(key string, def int64) int64 {
	if val, ok := section[key]; ok {
		if i, err := utils.ParseByte(val); err == nil {
			return i
		}
	}
	return def
}

func (section Section) Bool(key string, def bool) bool {
	if val, ok := section[key]; ok {
		switch strings.ToLower(val) {
		case "false", "off", "disable", "0", "no", "null", "nil", "undefined":
			return false
		default:
			return true
		}
	}
	return def
}
