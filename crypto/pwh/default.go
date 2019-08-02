package pwh

var defaultPWH = New("...bla.bla.bla...", 1024)

func Config(publicSalt string, complexity int) {
	defaultPWH.Config(publicSalt, complexity)
}

func Hash(word string, salt string) string {
	return defaultPWH.Hash(word, salt)
}

func Match(word string, salt string, hash string) bool {
	return defaultPWH.Match(word, salt, hash)
}
