package pwh

import (
	"runtime"
	"testing"
)

var pw = "password"
var ps = "private_salt_key"

func TestPWH(t *testing.T) {
	hashes := map[string]int{}
	for i := 0; i < 1024; i++ {
		h := Hash(pw, ps)
		if !Match(pw, ps, h) {
			t.Fatal("match failed:", h, "->", pw)
		}
		hashes[h]++
	}
	t.Logf("%d results", len(hashes))
	for hash, times := range hashes {
		t.Logf("%s(%d)", hash, times)
	}
}

func BenchmarkMatch(b *testing.B) {
	b.StopTimer()
	hs := map[int]string{}
	for i := 0; i < b.N; i++ {
		hs[i] = Hash(pw, ps)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !Match(pw, ps, hs[i]) {
			b.Fatal("match failed:", hs[i], "->", pw)
		}
	}
}

func BenchmarkMatchX(b *testing.B) {
	b.StopTimer()
	hs := map[int]string{}
	for i := 0; i < b.N; i++ {
		hs[i] = Hash(pw, ps)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !MatchX(pw, ps, hs[i], runtime.NumCPU()) {
			b.Fatal("match failed:", hs[i], "->", pw)
		}
	}
}
