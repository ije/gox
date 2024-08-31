package valid

import (
	"strings"

	"github.com/ije/gox/utils"
)

var (
	r0_9       = Range{'0', '9'}
	ra_z       = Range{'a', 'z'}
	rA_Z       = Range{'A', 'Z'}
	vNum       = Validator{r0_9}
	vHex       = Validator{r0_9, Range{'a', 'f'}, Range{'A', 'F'}}
	vSlug      = Validator{r0_9, ra_z, rA_Z, Eq('-')}
	vEmailName = Validator{r0_9, ra_z, rA_Z, Eq('.'), Eq('-'), Eq('_'), Eq('+')}
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
	return !startsWithAny(s, '-') && vSlug.Is(s)
}

func IsDomain(s string) bool {
	for {
		i := strings.LastIndexByte(s, '.')
		if i == -1 {
			return vSlug.Is(s)
		}
		if !vSlug.Is(s[i+1:]) {
			return false
		}
		s = s[:i]
	}
}

func IsEmail(s string) bool {
	// minimal email: a@b.cn
	if len(s) < 6 {
		return false
	}

	name, domain := utils.SplitByLastByte(s, '@')
	return !startsWithAny(name, '.', '-', '_', '+') && vEmailName.Is(name) && IsDomain(domain)
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

func startsWithAny(s string, cs ...byte) bool {
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
