package pwh

import (
	"crypto/sha512"
)

const pwTable = "*?0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var globalPWHasher *PWHasher

func Config(publicSalt string, complexity int) {
	if complexity < 1 {
		complexity = 1
	}
	saltHasher := sha512.New()
	saltHasher.Write([]byte(publicSalt))
	globalPWHasher.complexity = complexity
	globalPWHasher.publicSalt = saltHasher.Sum(nil)
}

func Hash(word, salt string) string {
	return globalPWHasher.Hash(word, salt)
}

func Match(word, salt, hash string) bool {
	return globalPWHasher.Match(word, salt, hash)
}

func MatchX(word, salt, hash string, routines int) bool {
	return globalPWHasher.MatchX(word, salt, hash, routines)
}

func init() {
	globalPWHasher = New("gox", 1024)
}
