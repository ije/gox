package pwh

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"math/rand"
	"sync"
)

const pwTable = "*?0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type PWHasher struct {
	lock       sync.RWMutex
	publicSalt []byte
	complexity int
}

func New(publicSalt string, complexity int) (pwh *PWHasher) {
	pwh = &PWHasher{}
	pwh.Config(publicSalt, complexity)
	return
}

func (pwh *PWHasher) Config(publicSalt string, complexity int) {
	if complexity < 1 {
		complexity = 1
	}
	publicSaltHasher := sha512.New()
	publicSaltHasher.Write([]byte(publicSalt))

	pwh.lock.Lock()
	defer pwh.lock.Unlock()

	pwh.complexity = complexity
	pwh.publicSalt = publicSaltHasher.Sum(nil)
}

func (pwh *PWHasher) Hash(word, salt string) string {
	pwh.lock.RLock()
	defer pwh.lock.RUnlock()

	return string(pwh.hash(rand.Int()%pwh.complexity, word, salt))
}

func (pwh *PWHasher) Match(word, salt, hash string) bool {
	pwh.lock.RLock()
	defer pwh.lock.RUnlock()

	for i := 0; i < pwh.complexity; i++ {
		if bytes.Equal([]byte(hash), pwh.hash(i, word, salt)) {
			return true
		}
	}
	return false
}

func (pwh *PWHasher) MatchX(word, salt, hash string, routines int) bool {
	if routines < 2 {
		return pwh.Match(word, salt, hash)
	}

	pwh.lock.RLock()
	defer pwh.lock.RUnlock()

	groups := (pwh.complexity + routines - 1) / routines
	matchc := make(chan bool, routines)
	matched := 0
	for i := 0; i < routines; i++ {
		go func(i int) {
			for s, e := i*groups, (i+1)*groups; s < e; s++ {
				if matched == routines {
					return
				}
				if bytes.Equal([]byte(hash), pwh.hash(s, word, salt)) {
					matchc <- true
					return
				}
			}
			matchc <- false
		}(i)
	}
	for {
		if <-matchc {
			matched = routines // Use for stoping all hasher goroutines
			return true
		} else if matched++; matched == routines {
			return false
		}
	}
}

func (pwh *PWHasher) hash(r int, word, salt string) []byte {
	codeTable := make([]byte, 64)
	for i, p := 0, rand.New(rand.NewSource(int64(r))).Perm(64); i < 64; i++ {
		codeTable[i] = pwTable[p[i]]
	}
	hmac := hmac.New(sha512.New384, pwh.publicSalt)
	hmac.Write([]byte(word))
	hmac.Write([]byte(salt))
	hashBytes := hmac.Sum(nil)
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

var defaultPWHasher *PWHasher

func Config(publicSalt string, complexity int) {
	if complexity < 1 {
		complexity = 1
	}
	publicSaltHasher := sha512.New()
	publicSaltHasher.Write([]byte(publicSalt))
	defaultPWHasher.complexity = complexity
	defaultPWHasher.publicSalt = publicSaltHasher.Sum(nil)
}

func Hash(word, salt string) string {
	return defaultPWHasher.Hash(word, salt)
}

func Match(word, salt, hash string) bool {
	return defaultPWHasher.Match(word, salt, hash)
}

func MatchX(word, salt, hash string, routines int) bool {
	return defaultPWHasher.MatchX(word, salt, hash, routines)
}

func init() {
	defaultPWHasher = New("go", 1024)
}
