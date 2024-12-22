package pwh

var defaultPWH = New([]byte("2006-03-04 15:06:07"), 10)

func Config(publicSalt []byte, cost int) {
	defaultPWH.Config(publicSalt, cost)
}

func Hash(word []byte, salt []byte) []byte {
	return defaultPWH.Hash(word, salt)
}

func Match(word []byte, salt []byte, hash []byte) bool {
	return defaultPWH.Match(word, salt, hash)
}
