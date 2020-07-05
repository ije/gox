package pwh

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"math/rand"
	"time"
)

const pwTable = "*?0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type PWHasher struct {
	publicSalt []byte
	cost       int
}

func New(publicSalt string, cost int) (pwh *PWHasher) {
	pwh = &PWHasher{}
	pwh.Config(publicSalt, cost)
	return
}

func (pwh *PWHasher) Config(publicSalt string, cost int) {
	if cost < 1 {
		cost = 1
	}
	saltHasher := sha512.New()
	saltHasher.Write([]byte(publicSalt))

	pwh.publicSalt = saltHasher.Sum(nil)
	pwh.cost = cost
}

func (pwh *PWHasher) Hash(word string, salt string) string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(int64(pwh.cost))
	return string(pwh.hash(seed, word, salt))
}

func (pwh *PWHasher) Match(word string, salt string, hash string) bool {
	for i := 0; i < pwh.cost; i++ {
		if bytes.Equal([]byte(hash), pwh.hash(int64(i), word, salt)) {
			return true
		}
	}
	return false
}

func (pwh *PWHasher) hash(seed int64, word string, salt string) []byte {
	codeTable := make([]byte, 64)
	for i, p := 0, rand.New(rand.NewSource(seed)).Perm(64); i < 64; i++ {
		codeTable[i] = pwTable[p[i]]
	}
	hasher := hmac.New(sha512.New, pwh.publicSalt)
	hasher.Write([]byte(salt))
	hasher = hmac.New(sha512.New384, hasher.Sum(nil))
	hasher.Write([]byte(word))
	hashBytes := hasher.Sum(nil)
	hash := make([]byte, 64)
	for i, j := 0, 0; i < 48; i += 3 {
		j = i * 4 / 3
		hash[j] = codeTable[hashBytes[i]>>2]
		hash[j+1] = codeTable[(hashBytes[i]&0x3)<<4|hashBytes[i+1]>>4]
		hash[j+2] = codeTable[(hashBytes[i+1]&0xf)<<2|hashBytes[i+2]>>6]
		hash[j+3] = codeTable[hashBytes[i+2]&0x3f]
	}
	return hash
}
