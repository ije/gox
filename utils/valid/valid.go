package valid

import (
	"regexp"
)

var (
	regIsEmail  = regexp.MustCompile(`^(?i)[0-9a-z]+([\.\-\_][0-9a-z]+)*@[0-9a-z]+([\-\.][0-9a-z]+)*?(\.[a-z]{2,})+$`)
	regIsNumber = regexp.MustCompile(`^\d+$`)
)

func IsEmail(s string) bool {
	return regIsEmail.MatchString(s)
}

func IsNumber(s string) bool {
	return regIsNumber.MatchString(s)
}

func IsHexString(s string, l int) bool {
	sl := len(s)
	if sl == 0 || (l > 0 && sl != l) {
		return false
	}
	for i := 0; i < sl; i++ {
		if c := s[i]; c < '0' || (c > '9' && c < 'A') || (c > 'F' && c < 'a') || c > 'f' {
			return false
		}
	}
	return true
}
