package valid

type Validator []Range

func (v Validator) Is(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		inRange := false
		for _, r := range v {
			if r.In(c) {
				inRange = true
				break
			}
		}
		if !inRange {
			return false
		}
	}

	return true
}
