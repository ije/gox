package valid

import (
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
	// minimal email: a@b.cn
	if len(s) < 6 {
		return false
	}

	name, domain := utils.SplitByLastByte(s, '@')
	return !hasAnyfix(name, '.', '-', '_', '+') && vEmailName.Is(name) && IsDomain(domain)
}

func IsIP(s string) bool {
	return IsIPv4(s) || IsIPv6(s)
}

func IsIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}

	for _, p := range parts {
		l := len(p)
		if !vNum.Is(p) || l > 3 {
			return false
		}
		if l == 3 && (p[0] > '2' || (p[0] == '2' && (p[1] > '5' || p[2] > '5'))) {
			return false
		}
	}

	return true
}

func IsIPv6(s string) bool {
	// todo: implement IsIPv6
	return false
}

func hasAnyfix(s string, cs ...byte) bool {
	l := len(s)
	if l == 0 {
		return false
	}

	for _, c := range cs {
		if c == s[0] || (l > 1 && c == s[l-1]) {
			return true
		}
	}

	return false
}
