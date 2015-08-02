/*

Config Package.

	package main

	import "github.com/ije/gox/config"

	func main() {
	    conf, err := config.New("a.conf")
	    if err != nil {
			return
	    }
	    log.Printf(conf.String("key", "defaultValue"))
	    log.Printf(conf.Section("sectionName").String("key", "defaultValue"))
	}

*/
package config

import (
	"bytes"
	"io"
	"os"
	"regexp"
)

type Config struct {
	defaultSection   Section
	extendedSections map[string]Section
}

func New(configFile string) (config *Config, err error) {
	config = &Config{}
	if len(configFile) > 0 {
		var file *os.File
		file, err = os.Open(configFile)
		if err != nil {
			if os.IsExist(err) {
				config = nil
			}
			return
		}
		defer file.Close()
		config.defaultSection, config.extendedSections, err = Parse(file)
	}
	return
}

func Parse(r io.Reader) (defaultSection Section, extendedSections map[string]Section, err error) {
	var n int
	var c byte
	var sectionKey string
	var section Section
	regSplitKV := regexp.MustCompile(`^([^ ]+)\s+(.+)$`)
	regSplitKVWithLongKey := regexp.MustCompile(`^"([^"]+)"\s+(.+)$`)
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
				matches := regSplitKV.FindSubmatch(line)
				if len(matches) == 3 {
					section[string(matches[1])] = string(matches[2])
				} else {
					section[string(line)] = ""
				}
			}
		}
	}
	buf := make([]byte, 1)
	line := bytes.NewBuffer(nil)

	section = Section{}
	extendedSections = map[string]Section{}

	for {
		n, err = r.Read(buf)
		if err != nil {
			if err != io.EOF {
				return
			}
			err = nil
			break
		}
		if n == 0 {
			break
		}
		c = buf[0]
		if c == '\r' || c == '\n' {
			parse(line.Bytes())
			line.Reset()
		} else {
			line.WriteByte(c)
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

func (config *Config) Bool(key string, def bool) bool {
	return config.defaultSection.Bool(key, def)
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

func (config *Config) Float64(key string, def float64) float64 {
	return config.defaultSection.Float64(key, def)
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

func (config *Config) ExtendedSections() map[string]Section {
	return config.extendedSections
}
