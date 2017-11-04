package session

import (
	"fmt"
	"strings"
	"time"

	"github.com/ije/gox/utils"
)

type Session struct {
	SID     string
	Values  map[string]interface{}
	Expires time.Time
}

func (session *Session) Has(key string) (ok bool) {
	_, ok = session.Values[key]
	return
}

func (session *Session) Bool(key string) (value bool) {
	if v, ok := session.Values[key]; ok {
		if b, ok := v.(bool); ok {
			value = b
		} else if f64, ok := utils.ToNumber(v); ok && f64 != 0 {
			value = true
		} else if s, ok := v.(string); ok && strings.ToLower(s) != "false" && s != "0" {
			value = true
		}
	}
	return
}

func (session *Session) String(key string) (value string) {
	if v, ok := session.Values[key]; ok {
		if s, ok := v.(string); ok {
			value = s
		} else {
			value = fmt.Sprintf("%v", v)
		}
	}
	return
}

func (session *Session) Number(key string) (value float64) {
	if v, ok := session.Values[key]; ok {
		value, ok = utils.ToNumber(v)
	}
	return
}

func (session *Session) Int(key string) (value int) {
	return int(session.Number(key))
}

func (session *Session) Set(key string, value interface{}) {
	if session.Values == nil {
		session.Values = map[string]interface{}{}
	}
	session.Values[key] = value
}
