package valid

type V interface {
	Match(c rune) bool
}

type Validator []V

func (v Validator) Match(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		matched := false
		for _, r := range v {
			if r.Match(c) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}
