package valid

import (
	"strconv"
	"strings"

	"github.com/ije/gox/utils"
)

var (
	vReg_09   = &Validator{{'0', '9'}}
	vReg_az   = &Validator{{'a', 'z'}}
	vReg_09AZ = &Validator{{'0', '9'}, {'A', 'Z'}}
	vReg_w    = &Validator{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}}
	vHex      = &Validator{{'0', '9'}, {'a', 'f'}, {'A', 'F'}}
	vDomain   = &Validator{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}}
	vSlug     = &Validator{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}, {'_', 0}}
)

func IsNumber(s string, a ...int) bool {
	for i, p := range strings.Split(s, ".") {
		if i > 1 || !vReg_09.Is(p) {
			return false
		}
	}

	return true
}

func IsHexString(s string, a ...int) bool {
	return vHex.Is(s, a...)
}

func IsIETFLangTag(s string) bool {
	l, c := utils.SplitByFirstByte(s, '-')
	if !vReg_az.Is(l, 2) {
		return false
	}

	if len(c) > 0 {
		return vReg_09AZ.Is(c)
	}

	return true
}

func IsIP(s string) bool {
	return IsIPv4(s) || IsIPv6(s)
}

func IsIPv4(s string) bool {
	for i, p := range strings.Split(s, ".") {
		if i > 3 || !vReg_09.Is(p, 1, 3) {
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

func IsDomain(s string) bool {
	if len(s) < 4 {
		return false
	}

	dn, dt := utils.SplitByLastByte(s, '.')
	return vDomain.Is(dn) && vReg_w.Is(dt)
}

func IsSlug(s string, a ...int) bool {
	l := len(s)
	if l == 0 {
		return false
	}

	for _, c := range []byte{s[0], s[l-1]} {
		switch c {
		case '.', '-':
			return false
		}
	}

	return vSlug.Is(s, a...)
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

	return vSlug.Is(name)
}
