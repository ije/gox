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

type Validator []Range

func (v Validator) Is(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		inrange := false
		for _, r := range v {
			if r.In(c) {
				inrange = true
				break
			}
		}
		if !inrange {
			return false
		}
	}

	return true
}
