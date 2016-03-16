package pwh

var globalPWHasher *PWHasher

func Config(publicSalt string, complexity int) {
	globalPWHasher.Config(publicSalt, complexity)
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
