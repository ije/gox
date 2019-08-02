package pwh

import (
	"testing"
)

var password = "password"
var salt = "private_salt"
var publicSalt = "public_salt"
var testHasher = New(publicSalt, 5120)

func TestPWH(t *testing.T) {
	hashes := map[string]int{}
	for i := 0; i < 100; i++ {
		h := testHasher.Hash(password, salt)
		if !testHasher.Match(password, salt, h) {
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

func BenchmarkMatch(b *testing.B) {
	b.StopTimer()
	hashes := map[int]string{}
	for i := 0; i < b.N; i++ {
		hashes[i] = testHasher.Hash(password, salt)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if !testHasher.Match(password, salt, hashes[i]) {
			b.Fatal("match failed:", hashes[i], "->", password)
		}
	}
}
