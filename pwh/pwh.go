package pwh

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"math/rand"
	"strings"

	"github.com/ije/go/utils/set"
)

const defaultPWTable = "./0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type PWHasher struct {
	pwTable    [64]byte
	publicSalt []byte
	complexity int
}

func New(publicSalt string, complexity int, pwTables ...string) *PWHasher {
	var pwTable [64]byte
	set := set.New()
	for _, c := range []byte(strings.Join(pwTables, "")) {
		set.Add(c)
	}
	if len(set) >= 64 {
		for i, v := range set.List()[:64] {
			pwTable[i] = v.(byte)
		}
	} else {
		for i := 0; i < 64; i++ {
			pwTable[i] = defaultPWTable[i]
		}
	}
	publicSaltHasher := sha512.New()
	publicSaltHasher.Write([]byte(publicSalt))
	if complexity < 1 {
		complexity = 1
	}
	return &PWHasher{pwTable, publicSaltHasher.Sum(nil), complexity}
}

func (pwh *PWHasher) Hash(word, salt string) string {
	return string(pwh.hash(rand.Int()%pwh.complexity, word, salt))
}

func (pwh *PWHasher) Match(word, salt, hash string) bool {
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
	tasks := (pwh.complexity + routines - 1) / routines
	matchc := make(chan bool, routines)
	matched := 0
	for i := 0; i < routines; i++ {
		go func(i int) {
			for s, e := i*tasks, (i+1)*tasks; s < e; s++ {
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
		select {
		case ok := <-matchc:
			if ok {
				matched = routines // Use for stoping all hasher goroutines
				return true
			} else if matched++; matched == routines {
				return false
			}
		}
	}
}

func (pwh *PWHasher) hash(seed int, word, salt string) []byte {
	codeTable := make([]byte, 64)
	for i, p := 0, rand.New(rand.NewSource(int64(seed))).Perm(64); i < 64; i++ {
		codeTable[i] = pwh.pwTable[p[i]]
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
