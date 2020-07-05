package pwh

var defaultPWH = New("gox", 10)

func Config(publicSalt string, cost int) {
	defaultPWH.Config(publicSalt, cost)
}

func Hash(word string, salt string) string {
	return defaultPWH.Hash(word, salt)
}

func Match(word string, salt string, hash string) bool {
	return defaultPWH.Match(word, salt, hash)
}
