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

func New(publicSalt []byte, cost int) (pwh *PWHasher) {
	pwh = &PWHasher{}
	pwh.Config(publicSalt, cost)
	return
}

func (pwh *PWHasher) Config(publicSalt []byte, cost int) {
	saltHasher := sha512.New()
	saltHasher.Write(publicSalt)

	if cost < 0 {
		cost = 0
	} else if cost > 20 {
		cost = 20
	}

	pwh.publicSalt = saltHasher.Sum(nil)
	pwh.cost = 1 << cost
}

func (pwh *PWHasher) Hash(password []byte, salt []byte) []byte {
	seed := rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(int64(pwh.cost))
	return pwh.hash(seed, password, salt)
}

func (pwh *PWHasher) Match(word []byte, salt []byte, hash []byte) bool {
	for i := 0; i < pwh.cost; i++ {
		if bytes.Equal(hash, pwh.hash(int64(i), word, salt)) {
			return true
		}
	}
	return false
}

func (pwh *PWHasher) hash(seed int64, password []byte, salt []byte) []byte {
	codeTable := make([]byte, 64)
	for i, p := 0, rand.New(rand.NewSource(seed)).Perm(64); i < 64; i++ {
		codeTable[i] = pwTable[p[i]]
	}
	h := hmac.New(sha512.New, salt)
	h.Write(password)
	h2 := hmac.New(sha512.New384, pwh.publicSalt)
	h2.Write(h.Sum(nil))
	b := h2.Sum(nil)
	hash := make([]byte, 64)
	for i, j := 0, 0; i < 48; i += 3 {
		j = i * 4 / 3
		hash[j] = codeTable[b[i]>>2]
		hash[j+1] = codeTable[(b[i]&0x3)<<4|b[i+1]>>4]
		hash[j+2] = codeTable[(b[i+1]&0xf)<<2|b[i+2]>>6]
		hash[j+3] = codeTable[b[i+2]&0x3f]
	}
	return hash
}
