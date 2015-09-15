package session

import (
	"net/http"
	"time"

	"github.com/ije/gox/cache"
	"github.com/ije/gox/crypto/mist"
)

type Session struct {
	prevExpired bool
	storage     cache.Cache
	lifetime    time.Duration
	values      map[string]interface{}
	http.ResponseWriter
	*http.Cookie
}

func Init(storage cache.Cache, w http.ResponseWriter, cookie *http.Cookie, lifetime time.Duration) (sess *Session, err error) {
	if cookie == nil || !sidValid(cookie.Value) {
		if cookie != nil {
			cookie = &http.Cookie{Name: cookie.Name, Path: cookie.Path, Domain: cookie.Domain}
		} else {
			cookie = &http.Cookie{Name: "session"}
		}
	SIDGEN:
		cookie.Value = sidGen()
		var ok bool
		ok, err = storage.Has(cookie.Value)
		if err != nil {
			return
		} else if ok {
			goto SIDGEN
		}
		cookie.HttpOnly = true
		cookie.Expires = time.Now().Add(lifetime)
		w.Header().Add("Set-Cookie", cookie.String())
	}

	sess = &Session{storage: storage, values: map[string]interface{}{}, lifetime: lifetime, ResponseWriter: w, Cookie: cookie}
	v, err := storage.Get(sess.Value)
	if err != nil {
		if err == cache.ErrNotFound || err.Error() == cache.ErrNotFound.Error() {
			err = nil
		} else if err == cache.ErrExpired || err.Error() == cache.ErrExpired.Error() {
			err = nil
			sess.prevExpired = true
		}
		return
	}

	values, ok := v.(map[string]interface{})
	if ok && values != nil {
		sess.values = values
	}

	if !ok {
		if err = storage.Delete(sess.Value); err != nil {
			sess = nil
			return
		}
	}
	return
}

func (sess *Session) SID() string {
	return sess.Value
}

func (sess *Session) Get(key string) (value interface{}) {
	value, _ = sess.values[key]
	return
}

func (sess *Session) Add(key string, value interface{}) {
	if _, ok := sess.values[key]; !ok {
		sess.values[key] = value
	}
}

func (sess *Session) Set(key string, value interface{}) {
	sess.values[key] = value
}

func (sess *Session) Delete(key string) {
	delete(sess.values, key)
}

func (sess *Session) Save() error {
	return sess.storage.Set(sess.Value, sess.values, sess.lifetime)
}

func (sess *Session) Destory() error {
	sess.Value = "-"
	sess.Expires = time.Now().Truncate(time.Second).UTC()
	sess.HttpOnly = true
	sess.ResponseWriter.Header().Add("Set-Cookie", sess.String())
	return sess.storage.Delete(sess.Value)
}

func sidValid(sid string) bool {
	if len(sid) != 64 {
		return false
	}
	for i := 0; i < 64; i++ {
		if c := sid[i]; c < '.' || (c > '9' && c < 'A') || (c > 'Z' && c < 'a') || c > 'z' {
			return false
		}
	}
	return true
}

func sidGen() string {
	return mist.Base64.String(64)
}
