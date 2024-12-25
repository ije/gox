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

// IsDigtalOnlyString returns true if the string s contains only digits.
func IsDigtalOnlyString(s string) bool {
	for _, c := range s {
		if !r0_9.Match(c) {
			return false
		}
	}
	return true
}

// IsHexString returns true if the string s is a valid hex string.
func IsHexString(s string) bool {
	return vHex.Match(s)
}

// IsNumber returns true if the string s is a valid number.
func IsNumber(s string) bool {
	if len(s) > 1 && s[0] == '-' {
		s = s[1:]
	}
	i, f := utils.SplitByLastByte(s, '.')
	return vNum.Match(i) && (f == "" || vNum.Match(f))
}

// IsSlug returns true if the string s is a valid slug.
func IsSlug(s string) bool {
	return !between(s, '-') && vSlug.Match(s)
}

// IsDomain returns true if the string s is a valid domain.
func IsDomain(s string) bool {
	for {
		i := strings.LastIndexByte(s, '.')
		if i == -1 {
			return vSlug.Match(s)
		}
		if !vSlug.Match(s[i+1:]) {
			return false
		}
		s = s[:i]
	}
}

// IsEmail returns true if the string s is a valid email.
func IsEmail(s string) bool {
	if len(s) < 3 {
		return false
	}
	name, domain := utils.SplitByLastByte(s, '@')
	if name == "" || domain == "" {
		return false
	}
	return !between(name, '.', '-', '_', '+') && vEmailName.Match(name) && IsDomain(domain)
}

// IsIPv4 returns true if the string s is a valid IPv4 address.
func IsIPv4(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return false
	}

	for _, p := range parts {
		l := len(p)
		if !vNum.Match(p) || l > 3 {
			return false
		}
		if l == 3 && (p[0] > '2' || (p[0] == '2' && (p[1] > '5' || p[2] > '5'))) {
			return false
		}
	}

	return true
}

func between(s string, bytes ...byte) bool {
	l := len(s)
	if l == 0 {
		return false
	}

	for _, c := range bytes {
		if c == s[0] || (l > 1 && c == s[l-1]) {
			return true
		}
	}

	return false
}
