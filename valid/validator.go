package valid

type Range struct {
	From byte
	To   byte
}

type Validator []Range

func (v Validator) Is(s string) bool {
	l := len(s)
	if l == 0 {
		return false
	}

	for l--; l >= 0; l-- {
		if !v.inRanges(s[l]) {
			return false
		}
	}

	return true
}

func (v Validator) inRanges(c byte) bool {
	for _, r := range v {
		if r.To > 0 {
			if c >= r.From && c <= r.To {
				return true
			}
		} else if c == r.From {
			return true
		}
	}
	return false
}
