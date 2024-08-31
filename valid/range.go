package valid

type Matcher interface {
	Match(c rune) bool
}

type Eq rune

func (r Eq) Match(c rune) bool {
	return rune(r) == c
}

type Range [2]rune

func (r Range) Match(c rune) bool {
	return c >= r[0] && c <= r[1]
}
