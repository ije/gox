package valid

import (
	"strings"

	"github.com/ije/gox/utils"
)

var (
	v09        = &Validator{[]Range{{'0', '9'}}}
	vaz        = &Validator{[]Range{{'a', 'z'}}}
	vazAZ      = &Validator{[]Range{{'a', 'z'}, {'A', 'Z'}}}
	v09AZ      = &Validator{[]Range{{'0', '9'}, {'A', 'Z'}}}
	vHex       = &Validator{[]Range{{'0', '9'}, {'a', 'f'}, {'A', 'F'}}}
	vSlug      = &Validator{[]Range{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}}}
	vEmailName = &Validator{[]Range{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}, {'_', 0}}}
)

func IsNumber(s string, n ...int) bool {
	return v09.Is(s, n...)
}

func IsHexString(s string, n ...int) bool {
	return vHex.Is(s, n...)
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
	for i, p := range strings.Split(s, ".") {
		if i > 3 || !v09.Is(p, 1, 3) || p[0] > '2' {
			return false
		}
	}
	return true
}

func IsDomain(s string) bool {
	if len(s) < 4 {
		return false
	}
	dn, dt := utils.SplitByLastByte(s, '.')
	return IsSlug(dn, 0) && vazAZ.Is(dt)
}

func IsSlug(s string, maxLen int) bool {
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
