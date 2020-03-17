package valid

import (
	"strconv"
	"strings"

	"github.com/ije/gox/utils"
)

var (
	rNum       = FromTo{'0', '9'}
	rword      = FromTo{'a', 'z'}
	rWORD      = FromTo{'A', 'Z'}
	vNum       = Validator{rNum}
	vHex       = Validator{rNum, FromTo{'a', 'f'}, FromTo{'A', 'F'}}
	vSlug      = Validator{rNum, rword, rWORD, Eq('-')}
	vEmailName = Validator{rNum, rword, rWORD, Eq('.'), Eq('-'), Eq('_'), Eq('+')}
)

func IsNumber(s string) bool {
	if len(s) > 1 && s[0] == '-' {
		s = s[1:]
	}
	integer, floater := utils.SplitByLastByte(s, '.')
	return vNum.Is(integer) && (floater == "" || vNum.Is(floater))
}

func IsHexString(s string) bool {
	return vHex.Is(s)
}

func IsIP(s string) bool {
	return IsIPv4(s) || IsIPv6(s)
}

func IsIPv4(s string) bool {
	for i, p := range strings.Split(s, ".") {
		if i > 3 || !vNum.Is(p) || len(p) > 3 {
			return false
		}
		if i, _ := strconv.Atoi(p); i > 255 {
			return false
		}
	}

	return true
}

func IsIPv6(s string) bool {
	return false
}

func IsSlug(s string) bool {
	return !hasAnyfix(s, '-') && vSlug.Is(s)
}

func IsDomain(s string) bool {
	for _, p := range strings.Split(s, ".") {
		if !IsSlug(p) {
			return false
		}
	}
	return true
}

func IsEmail(s string) bool {
	if len(s) < 6 {
		return false
	}

	name, domain := utils.SplitByLastByte(s, '@')
	return !hasAnyfix(name, '.', '-', '_', '+') && vEmailName.Is(name) && IsDomain(domain)
}

func hasAnyfix(s string, cs ...byte) bool {
	l := len(s)
	if l == 0 {
		return false
	}

	for _, c := range []byte{s[0], s[l-1]} {
		for _, _c := range cs {
			if c == _c {
				return true
			}
		}
	}

	return false
}
