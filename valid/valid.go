package valid

import (
	"strconv"
	"strings"

	"gox/utils"
)

var (
	r09        = FromTo{'0', '9'}
	raz        = FromTo{'a', 'z'}
	rAZ        = FromTo{'A', 'Z'}
	v09        = Validator{r09}
	vaz        = Validator{raz}
	v09AZ      = Validator{r09, rAZ}
	v09azAZ    = Validator{r09, raz, rAZ}
	vHex       = Validator{r09, FromTo{'a', 'f'}, FromTo{'A', 'F'}}
	vSlug      = Validator{r09, raz, rAZ, Eq('.'), Eq('-')}
	vEmailName = Validator{r09, raz, rAZ, Eq('.'), Eq('-'), Eq('_'), Eq('+')}
)

func IsNumber(s string) bool {
	if len(s) > 1 && s[0] == '-' {
		s = s[1:]
	}
	inter, floater := utils.SplitByLastByte(s, '.')
	return v09.Is(inter) && (floater == "" || v09.Is(floater))
}

func IsHexString(s string) bool {
	return vHex.Is(s)
}

func IsIETFLangTag(s string) bool {
	l, c := utils.SplitByFirstByte(s, '-')
	return len(l) == 2 && vaz.Is(l) && (len(c) == 0 || v09AZ.Is(c))
}

func IsIP(s string) bool {
	return IsIPv4(s) || IsIPv6(s)
}

func IsIPv4(s string) bool {
	for i, p := range strings.Split(s, ".") {
		if i > 3 || !v09.Is(p) || len(p) > 3 {
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
	return !hasPreSuffix(s, '.', '-') && vSlug.Is(s)
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
	return !hasPreSuffix(name, '.', '-', '_', '+') && vEmailName.Is(name) && IsDomain(domain)
}

func hasPreSuffix(s string, cs ...byte) bool {
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
