package pwh

import (
	"testing"

	"github.com/ije/gox/crypto/rand"
)

func TestPWH(t *testing.T) {
	hashes := map[string]int{}
	password := rand.Hex.Bytes(8)
	privateSalt := rand.Hex.Bytes(16)
	for i := 0; i < 100; i++ {
		h := Hash(password, privateSalt)
		if !Match(password, privateSalt, h) {
			t.Fatal("failed to match hash ", h, "->", password)
		}
		hashes[string(h)]++
	}

	for hash, repeatTimes := range hashes {
		if repeatTimes > 1 {
			t.Logf("[warn] repeated hash: %s (%d times)", hash, repeatTimes)
		}
	}
}

func BenchmarkMatchCost10(b *testing.B) {
	b.StopTimer()
	hashes := make([][3][]byte, b.N)
	hasher := New([]byte("."), 10)
	for i := 0; i < b.N; i++ {
		password := rand.Hex.Bytes(8)
		privateSalt := rand.Hex.Bytes(16)
		hashes[i] = [3][]byte{password, privateSalt, hasher.Hash(password, privateSalt)}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h := hashes[i]
		if !hasher.Match(h[0], h[1], h[2]) {
			b.Fatal("failed to match hash ", h[2], "->", h[0])
		}
	}
}

func BenchmarkMatchCost11(b *testing.B) {
	b.StopTimer()
	hashes := make([][3][]byte, b.N)
	hasher := New([]byte("."), 11)
	for i := 0; i < b.N; i++ {
		password := rand.Hex.Bytes(8)
		privateSalt := rand.Hex.Bytes(16)
		hashes[i] = [3][]byte{password, privateSalt, hasher.Hash(password, privateSalt)}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h := hashes[i]
		if !hasher.Match(h[0], h[1], h[2]) {
			b.Fatal("failed to match hash ", h[2], "->", h[0])
		}
	}
}

func BenchmarkMatchCost12(b *testing.B) {
	b.StopTimer()
	hashes := make([][3][]byte, b.N)
	hasher := New([]byte("."), 12)
	for i := 0; i < b.N; i++ {
		password := rand.Hex.Bytes(8)
		privateSalt := rand.Hex.Bytes(16)
		hashes[i] = [3][]byte{password, privateSalt, hasher.Hash(password, privateSalt)}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		h := hashes[i]
		if !hasher.Match(h[0], h[1], h[2]) {
			b.Fatal("failed to match hash ", h[2], "->", h[0])
		}
	}
}
