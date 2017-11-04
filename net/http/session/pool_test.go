package session

import (
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	pool := InitSessionPool(time.Second)
	sess := pool.Get("")
	t.Log(sess.SID)
	sess = pool.Get(sess.SID)
	t.Log(sess.SID)
	time.Sleep(time.Second)
	t.Log(sess.SID)
	time.Sleep(time.Second)
	sess = pool.Get(sess.SID)
	t.Log(sess.SID)
}
