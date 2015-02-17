/*

Config Package.

	package main

	import "github.com/ije/go/config"

	func main() {
	    conf, err := config.New("a.conf")
	    if err != nil {
			return
	    }
	    conf.String("key", "defaultValue")
	    conf.Section("sectionName").String("key", "defaultValue")
	}

*/
package config

import (
	"bytes"
	"io/ioutil"
	"regexp"
)

type Config struct {
	rawData          []byte
	defaultSection   Section
	extendedSections map[string]Section
}

func New(filename string) (config *Config, err error) {
	rwdata, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	config = &Config{rawData: rwdata}
	config.defaultSection, config.extendedSections = Parse(rwdata)
	return
}

func Parse(data []byte) (defaultSection Section, extendedSections map[string]Section) {
	var sectionKey string
	var section Section
	regSplitKV := regexp.MustCompile(`^([^ ]+)\s+(.+)$`)
	regSplitKVWithLongKey := regexp.MustCompile(`^"([^"]+)"\s*(.+)$`)
	parse := func(line []byte) {
		line = bytes.TrimSpace(line)
		if ll := len(line); ll > 0 {
			switch line[0] {
			case '#':
				return
			case '[':
				if ll >= 3 && line[ll-1] == ']' {
					if len(sectionKey) == 0 {
						defaultSection = section
					} else {
						extendedSections[sectionKey] = section
					}
					sectionKey = string(line[1 : ll-1])
					section = Section{}
					return
				}
			case '"':
				if ll >= 5 {
					matches := regSplitKVWithLongKey.FindSubmatch(line)
					if len(matches) == 3 {
						section[string(matches[1])] = string(matches[2])
					}
					return
				}
			default:
				if ll >= 3 {
					matches := regSplitKV.FindSubmatch(line)
					if len(matches) == 3 {
						section[string(matches[1])] = string(matches[2])
					}
				}
			}
		}
	}
	section = Section{}
	extendedSections = map[string]Section{}
	for i, j, l := 0, 0, len(data); i < l; i++ {
		switch data[i] {
		case '\r', '\n':
			if i > j {
				parse(data[j:i])
			}
			j = i + 1
		default:
			if i == l-1 && j < l {
				parse(data[j:])
			}
		}
	}
	if len(sectionKey) == 0 {
		defaultSection = section
	} else {
		extendedSections[sectionKey] = section
	}
	return
}

func (config *Config) IsEmpty() bool {
	return config.defaultSection.IsEmpty() && len(config.extendedSections) == 0
}

func (config *Config) Contains(key string) bool {
	return config.defaultSection.Contains(key)
}

func (config *Config) String(key, def string) string {
	return config.defaultSection.String(key, def)
}

func (config *Config) Int(key string, def int) int {
	return config.defaultSection.Int(key, def)
}

func (config *Config) Int64(key string, def int64) int64 {
	return config.defaultSection.Int64(key, def)
}

func (config *Config) Bytes(key string, def int64) int64 {
	return config.defaultSection.Bytes(key, def)
}

func (config *Config) Bool(key string, def bool) bool {
	return config.defaultSection.Bool(key, def)
}

func (config *Config) Set(key string, value interface{}) {
	config.defaultSection.Set(key, value)
}

func (config *Config) Section(name string) (section Section) {
	if len(name) == 0 {
		section = config.defaultSection
		return
	}
	section, ok := config.extendedSections[name]
	if ok {
		return
	}
	section = Section{}
	config.extendedSections[name] = section
	return
}
