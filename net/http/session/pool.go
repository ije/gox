package session

import (
	"sync"
	"time"

	"github.com/ije/gox/crypto/rs"
)

type SessionPool struct {
	lock            sync.RWMutex
	sessions        map[string]*Session
	sessionLifetime time.Duration
	gcTicker        *time.Ticker
}

func InitSessionPool(sessionLifetime time.Duration) *SessionPool {
	sp := &SessionPool{
		sessions: map[string]*Session{},
	}
	sp.SetSessionLifetime(sessionLifetime)
	go func(sp *SessionPool) {
		for {
			sp.lock.RLock()
			ticker := sp.gcTicker
			sp.lock.RUnlock()

			if ticker != nil {
				<-ticker.C
				sp.GC()
			}
		}
	}(sp)

	return sp
}

func (sp *SessionPool) SetSessionLifetime(lifetime time.Duration) {
	if lifetime < time.Second {
		return
	}

	sp.lock.Lock()
	defer sp.lock.Unlock()

	if sp.gcTicker != nil {
		sp.gcTicker.Stop()
	}
	sp.sessionLifetime = lifetime
	sp.gcTicker = time.NewTicker(lifetime)
}

func (sp *SessionPool) Get(sid string) (session *Session) {
	now := time.Now()
	ok := len(sid) == 64

	if ok {
		sp.lock.RLock()
		session, ok = sp.sessions[sid]
		sp.lock.RUnlock()
	}

	if ok && session.Expires.Before(now) {
		sp.lock.Lock()
		delete(sp.sessions, sid)
		sp.lock.Unlock()

		session = nil
	}

	if session == nil {
	NEWSID:
		sid = rs.Base64.String(64)
		sp.lock.RLock()
		_, ok := sp.sessions[sid]
		sp.lock.RUnlock()
		if ok {
			goto NEWSID
		}

		session = &Session{
			SID:    sid,
			Values: map[string]interface{}{},
		}
		sp.lock.Lock()
		sp.sessions[sid] = session
		sp.lock.Unlock()
	}

	sp.lock.Lock()
	session.Expires = now.Add(sp.sessionLifetime)
	sp.lock.Unlock()

	return
}

func (sp *SessionPool) Update(sid string, values map[string]interface{}) {
	sp.lock.RLock()
	sess, ok := sp.sessions[sid]
	sp.lock.RUnlock()

	if !ok {
		return
	}

	sp.lock.Lock()
	sess.Values = values
	sess.Expires = time.Now().Add(sp.sessionLifetime)
	sp.lock.Unlock()
	return
}

func (sp *SessionPool) Destroy(sid string) {
	sp.lock.Lock()
	defer sp.lock.Unlock()

	delete(sp.sessions, sid)
	return
}

func (sp *SessionPool) GC() {
	now := time.Now()

	sp.lock.RLock()
	defer sp.lock.RUnlock()

	for sid, session := range sp.sessions {
		if session.Expires.Before(now) {
			sp.lock.RUnlock()
			sp.lock.Lock()
			delete(sp.sessions, sid)
			sp.lock.Unlock()
			sp.lock.RLock()
		}
	}
}
