package valid

import (
	"strings"

	"github.com/ije/gox/utils"
)

var (
	v09        = &Validator{{'0', '9'}}
	vaz        = &Validator{{'a', 'z'}}
	vazAZ      = &Validator{{'a', 'z'}, {'A', 'Z'}}
	v09AZ      = &Validator{{'0', '9'}, {'A', 'Z'}}
	vHex       = &Validator{{'0', '9'}, {'a', 'f'}, {'A', 'F'}}
	vSlug      = &Validator{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}}
	vEmailName = &Validator{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}, {'_', 0}}
)

func IsNumber(s string, a ...int) bool {
	return v09.Is(s, a...)
}

func IsHexString(s string, a ...int) bool {
	return vHex.Is(s, a...)
}

func IsIETFLangTag(s string) bool {
	l, c := utils.SplitByFirstByte(s, '-')
	if !vaz.Is(l, 2) {
		return false
	}
	if len(c) > 0 {
		return v09AZ.Is(c)
	}
	return true
}

func IsIP(s string) bool {
	return IsIPv4(s) || IsIPv6(s)
}

func IsIPv4(s string) bool {
	for i, p := range strings.Split(s, ".") {
		if i > 3 || !v09.Is(p, 1, 3) || p[0] > '2' {
			return false
		}
	}
	return true
}

func IsIPv6(s string) bool {
	return false
}

func IsDomain(s string) bool {
	if len(s) < 4 {
		return false
	}
	dn, dt := utils.SplitByLastByte(s, '.')
	return IsSlug(dn, 0) && vazAZ.Is(dt)
}

func IsSlug(s string, a ...int) bool {
	var maxLen int
	if len(a) > 0 {
		maxLen = a[0]
	}
	l := len(s)
	if l == 0 || (maxLen > 0 && l > maxLen) {
		return false
	}
	for _, c := range []byte{s[0], s[l-1]} {
		switch c {
		case '.', '-':
			return false
		}
	}
	return vSlug.Is(s)
}

func IsEmail(s string) bool {
	if len(s) < 6 {
		return false
	}
	name, domain := utils.SplitByLastByte(s, '@')
	if !IsDomain(domain) {
		return false
	}
	nl := len(name)
	if nl == 0 {
		return false
	}
	for _, c := range []byte{name[0], name[nl-1]} {
		switch c {
		case '.', '_', '-':
			return false
		}
	}
	return vEmailName.Is(name)
}
