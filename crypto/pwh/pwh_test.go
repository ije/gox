package pwh

import (
	"testing"

	"github.com/ije/gox/crypto/rs"
)

const (
	publicSalt = "PUBLIC_SALT"
)

func TestPWH(t *testing.T) {
	hasher := New(publicSalt, 2048)
	hashes := map[string]int{}
	for i := 0; i < 100; i++ {
		password := rs.Hex.String(8)
		privateSalt := rs.Hex.String(16)
		h := hasher.Hash(password, privateSalt)
		if !hasher.Match(password, privateSalt, h) {
			t.Fatal("match failed:", h, "->", password)
		}
		hashes[h]++
	}

	for hash, repeatTimes := range hashes {
		if repeatTimes > 1 {
			t.Logf("[warn] repeat hash: %s (%d times)", hash, repeatTimes)
		}
	}
}

func BenchmarkMatchCost1024(b *testing.B) {
	b.StopTimer()
	hashes := map[int][3]string{}
	hasher := New(publicSalt, 1024)
	for i := 0; i < b.N; i++ {
		password := rs.Hex.String(8)
		privateSalt := rs.Hex.String(16)
		hashes[i] = [3]string{password, privateSalt, hasher.Hash(password, privateSalt)}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h := hashes[i]
		if !hasher.Match(h[0], h[1], h[2]) {
			b.Fatal("match failed:", h[2], "->", h[0])
		}
	}
}

func BenchmarkMatchCost5120(b *testing.B) {
	b.StopTimer()
	hashes := map[int][3]string{}
	hasher := New(publicSalt, 5120)
	for i := 0; i < b.N; i++ {
		password := rs.Hex.String(8)
		privateSalt := rs.Hex.String(16)
		hashes[i] = [3]string{password, privateSalt, hasher.Hash(password, privateSalt)}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h := hashes[i]
		if !hasher.Match(h[0], h[1], h[2]) {
			b.Fatal("match failed:", h[2], "->", h[0])
		}
	}
}
