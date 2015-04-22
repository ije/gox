package session

import (
	"net/http"
	"time"

	_cache "github.com/ije/gox/cache"
	"github.com/ije/gox/crypto/mist"
)

var (
	ErrNotFound = _cache.ErrNotFound
)

type Session struct {
	FromExpired bool
	cache       _cache.Cache
	lifetime    time.Duration
	values      map[string]interface{}
	http.ResponseWriter
	*http.Cookie
}

func InitSession(cache _cache.Cache, r *http.Request, w http.ResponseWriter, cookieName string, lifetime time.Duration) (sess *Session, err error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || !sidValid(cookie.Value) {
		cookie = &http.Cookie{}
	SIDGEN:
		cookie.Value = sidGen()
		var ok bool
		ok, err = cache.Has(cookie.Value)
		if err != nil {
			return
		} else if ok {
			goto SIDGEN
		}
		cookie.HttpOnly = true
		w.Header().Add("Set-Cookie", cookie.String())
	}

	sess = &Session{cache: cache, values: map[string]interface{}{}, lifetime: lifetime, ResponseWriter: w, Cookie: cookie}
	v, err := cache.Get(sess.Value)
	if err != nil {
		if err != _cache.ErrExpired {
			return
		}
		err = nil
		sess.FromExpired = true
		return
	}
	values, ok := v.(map[string]interface{})
	if !ok {
		err = cache.Delete(sess.Value)
		if err != nil {
			sess = nil
		}
	} else if values != nil {
		sess.values = values
	}
	return
}

func (sess *Session) SID() string {
	return sess.Value
}

func (sess *Session) Get(key string) (value interface{}, err error) {
	value, ok := sess.values[key]
	if !ok {
		err = ErrNotFound
	}
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
	return sess.cache.Set(sess.Value, sess.values, sess.lifetime)
}

func (sess *Session) Destory() (err error) {
	err = sess.cache.Delete(sess.Value)
	if err == nil {
		sess.Value = "deleted"
		sess.Expires = time.Now().Truncate(time.Second).UTC()
		sess.ResponseWriter.Header().Add("Set-Cookie", sess.String())
	}
	return
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
