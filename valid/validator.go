package valid

type Range struct {
	From byte
	To   byte
}

type Validator []Range

func (v Validator) Is(s string, a ...int) bool {
	l := len(s)
	if l == 0 {
		return false
	}

	if al := len(a); al == 1 {
		max := a[0]
		if max > 0 && l > max {
			return false
		}
	} else if al > 1 {
		min, max := a[0], a[1]
		if min > 0 || max > 0 {
			if min != max {
				if min > max {
					min, max = max, min
				}
				if l > max || (min > 0 && l < min) {
					return false
				}
			} else if min > 0 && l != min {
				return false
			}
		}
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
