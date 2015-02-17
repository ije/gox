package pwh

import (
	"runtime"
	"testing"
)

var pwh = New("public_salt_key", 1024)
var pw = "password"
var ps = "private_salt_key"

func TestPWH(t *testing.T) {
	for i := 0; i < 1024; i++ {
		h := pwh.Hash(pw, ps)
		t.Log(h, pwh.Match(pw, ps, h))
	}
}

func BenchmarkMatch(b *testing.B) {
	b.StopTimer()
	hs := map[int]string{}
	for i := 0; i < b.N; i++ {
		hs[i] = pwh.Hash(pw, ps)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pwh.Match(pw, ps, hs[i])
	}
}

func BenchmarkMatchX(b *testing.B) {
	b.StopTimer()
	hs := map[int]string{}
	for i := 0; i < b.N; i++ {
		hs[i] = pwh.Hash(pw, ps)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		pwh.MatchX(pw, ps, hs[i], runtime.NumCPU())
	}
}
