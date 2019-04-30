package valid

type Range interface {
	In(c rune) bool
}

type Eq rune

func (r Eq) In(c rune) bool {
	return rune(r) == c
}

type FromTo [2]rune

func (r FromTo) In(c rune) bool {
	return c >= r[0] && c <= r[1]
}
