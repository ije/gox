package pwh

import (
	"runtime"
	"testing"
)

var pwh = New("public_salt_key", 1024)
var pw = "hello"
var ps = "private_salt_key"

func TestPWH(t *testing.T) {
	for i := 0; i < 1024; i++ {
		h := pwh.Hash(pw, ps)
		if !pwh.Match(pw, ps, h) {
			t.Fatal("match failed:", h, "->", pw)
		}
		t.Log(h)
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
		if !pwh.Match(pw, ps, hs[i]) {
			b.Fatal("match failed:", hs[i], "->", pw)
		}
	}
}

func BenchmarkMatchX(b *testing.B) {
	b.StopTimer()
	hs := map[int]string{}
	for i := 0; i < b.N; i++ {
		hs[i] = pwh.Hash(pw, ps)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !pwh.MatchX(pw, ps, hs[i], runtime.NumCPU()) {
			b.Fatal("match failed:", hs[i], "->", pw)
		}
	}
}
