package valid

type Range struct {
	From byte
	To   byte
}

type Validator struct {
	ranges []Range
}

func (v *Validator) Is(s string, n ...int) bool {
	i := len(s)
	if i == 0 {
		return false
	}
	l := len(n)
	if l > 0 {
		min := n[0]
		if l == 1 && min > 0 && i != min {
			return false
		}
		if l > 1 {
			max := n[1]
			if min > max {
				min, max = max, min
			}
			if i < min || i > max {
				return false
			}
		}
	}
	for i--; i >= 0; i-- {
		if !v.inRanges(s[i]) {
			return false
		}
	}
	return true
}

func (v *Validator) inRanges(c byte) bool {
	for _, r := range v.ranges {
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
