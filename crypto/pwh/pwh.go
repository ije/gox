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
	complexity int
}

func New(publicSalt string, complexity int) (pwh *PWHasher) {
	pwh = &PWHasher{}
	pwh.config(publicSalt, complexity)
	return
}

func (pwh *PWHasher) config(publicSalt string, complexity int) {
	if complexity < 1 {
		complexity = 1
	}
	saltHasher := sha512.New()
	saltHasher.Write([]byte(publicSalt))

	pwh.complexity = complexity
	pwh.publicSalt = saltHasher.Sum(nil)
}

func (pwh *PWHasher) Hash(word, salt string) string {
	seed := rand.New(rand.NewSource(time.Now().UTC().UnixNano())).Int63n(int64(pwh.complexity))
	return string(pwh.hash(seed, word, salt))
}

func (pwh *PWHasher) Match(word, salt, hash string) bool {
	for i := 0; i < pwh.complexity; i++ {
		if bytes.Equal([]byte(hash), pwh.hash(int64(i), word, salt)) {
			return true
		}
	}
	return false
}

func (pwh *PWHasher) hash(seed int64, word, salt string) []byte {
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
