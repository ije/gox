package cache

import (
	"net/http"
	"time"

	"github.com/ije/go/crypto/mist"
)

type Session interface {
	SID() string
	Get(key string) (value interface{}, err error)
	Add(key string, value interface{})
	Set(key string, value interface{})
	Delete(key string)
	Save()
	Destory()
}

type session struct {
	httprw      http.ResponseWriter
	cache       Cache
	lifetime    time.Duration
	values      map[string]interface{}
	FromExpired bool
	*http.Cookie
}

func InitSession(r *http.Request, w http.ResponseWriter, cookieName string, lifetime time.Duration) Session {
	cache := New("go_session", lifetime)

	cookie, err := r.Cookie(cookieName)
	if err != nil || !sidValid(cookie.Value) {
		cookie = &http.Cookie{}
	SIDGEN:
		cookie.Value = sidGen()
		if cache.Has(cookie.Value) {
			goto SIDGEN
		}
		cookie.HttpOnly = true
		w.Header().Add("Set-Cookie", cookie.String())
	}

	sess := &session{cache: cache, values: map[string]interface{}{}, lifetime: lifetime, httprw: w, Cookie: cookie}
	v, err := cache.Get(sess.Value)
	if err != nil {
		sess.FromExpired = err == ErrExpired
		return sess
	}
	if values, ok := v.(map[string]interface{}); ok && values != nil {
		sess.values = values
	} else {
		cache.Delete(sess.Value)
	}
	return sess
}

func (sess *session) SID() string {
	return sess.Value
}

func (sess *session) Get(key string) (value interface{}, err error) {
	value, ok := sess.values[key]
	if !ok {
		err = ErrNotFound
	}
	return
}

func (sess *session) Add(key string, value interface{}) {
	if _, ok := sess.values[key]; !ok {
		sess.values[key] = value
	}
}

func (sess *session) Set(key string, value interface{}) {
	sess.values[key] = value
}

func (sess *session) Delete(key string) {
	delete(sess.values, key)
}

func (sess *session) Save() {
	sess.cache.Set(sess.Value, sess.values, sess.lifetime)
}

func (sess *session) Destory() {
	sess.cache.Delete(sess.Value)
	sess.Value = "deleted"
	sess.Expires = time.Now().Truncate(time.Second).UTC()
	sess.httprw.Header().Add("Set-Cookie", sess.String())
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
