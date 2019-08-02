package pwh

import (
	"testing"
)

const (
	password    = "password"
	privateSalt = "private_salt"
	publicSalt  = "public_salt"
)

func TestPWH(t *testing.T) {
	hasher := New(publicSalt, 1024)
	hashes := map[string]int{}
	for i := 0; i < 100; i++ {
		h := hasher.Hash(password, privateSalt)
		if !hasher.Match(password, privateSalt, h) {
			t.Fatal("match failed:", h, "->", password)
		}
		hashes[h]++
	}

	for hash, repeatTimes := range hashes {
		if repeatTimes > 1 {
			t.Logf("repeat hash: %s(%d times)", hash, repeatTimes)
		}
	}
}

func BenchmarkMatchCost1024(b *testing.B) {
	b.StopTimer()
	hashes := map[int]string{}
	hasher := New(publicSalt, 1024)
	for i := 0; i < b.N; i++ {
		hashes[i] = hasher.Hash(password, privateSalt)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !hasher.Match(password, privateSalt, hashes[i]) {
			b.Fatal("match failed:", hashes[i], "->", password)
		}
	}
}

func BenchmarkMatchCost5120(b *testing.B) {
	b.StopTimer()
	hashes := map[int]string{}
	hasher := New(publicSalt, 5120)
	for i := 0; i < b.N; i++ {
		hashes[i] = hasher.Hash(password, privateSalt)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !hasher.Match(password, privateSalt, hashes[i]) {
			b.Fatal("match failed:", hashes[i], "->", password)
		}
	}
}
