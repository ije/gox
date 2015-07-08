package valid

import (
	"github.com/ije/gox/utils"
	"strings"
)

var (
	emailNameValidator = &Validator{[]Range{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}, {'_', 0}}}
	DNValidator        = &Validator{[]Range{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}}}
	DTValidator        = &Validator{[]Range{{'a', 'z'}, {'A', 'Z'}}}
	numberValidator    = &Validator{[]Range{{'0', '9'}}}
	hexValidator       = &Validator{[]Range{{'0', '9'}, {'a', 'f'}, {'A', 'F'}}}
)

func IsNumber(s string, n ...int) bool {
	return numberValidator.Is(s, n...)
}

func IsHexString(s string, n ...int) bool {
	return hexValidator.Is(s, n...)
}

func IsIP(s string) bool {
	for i, p := range strings.Split(s, ".") {
		if i > 3 || !numberValidator.Is(p, 1, 3) || p[0] > '2' {
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
	dl, tl := len(dn), len(dt)
	if dl*tl == 0 {
		return false
	}
	for _, c := range []byte{dn[0], dn[dl-1]} {
		switch c {
		case '.', '-':
			return false
		}
	}
	if !DNValidator.Is(dn) || !DTValidator.Is(dt) {
		return false
	}
	return true
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
	if !emailNameValidator.Is(name) {
		return false
	}
	return true
}
