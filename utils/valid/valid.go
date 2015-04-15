package valid

import (
	"github.com/ije/aisling/utils"
)

var (
	emailNameValidator = &Validator{[]Range{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}, {'_', 0}}}
	emailDNValidator   = &Validator{[]Range{{'0', '9'}, {'a', 'z'}, {'A', 'Z'}, {'.', 0}, {'-', 0}}}
	emailDTValidator   = &Validator{[]Range{{'a', 'z'}, {'A', 'Z'}}}
	numberValidator    = &Validator{[]Range{{'0', '9'}}}
	hexValidator       = &Validator{[]Range{{'0', '9'}, {'a', 'f'}, {'A', 'F'}}}
)

func IsEmail(s string) bool {
	name, domain := utils.SplitByLastByte(s, '@')
	dn, dt := utils.SplitByLastByte(domain, '.')
	nl, dl, tl := len(name), len(dn), len(dt)
	if nl*dl*tl == 0 {
		return false
	}
	for _, c := range []byte{name[0], name[nl-1], dn[0], dn[dl-1]} {
		switch c {
		case '.', '_', '-':
			return false
		}
	}
	if !emailNameValidator.Is(name) || !emailDNValidator.Is(dn) || !emailDTValidator.Is(dt) {
		return false
	}
	return true
}

func IsNumber(s string, n ...int) bool {
	return numberValidator.Is(s, n...)
}

func IsHexString(s string, n ...int) bool {
	return hexValidator.Is(s, n...)
}
