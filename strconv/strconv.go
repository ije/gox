package strconv

import (
	"strconv"
)

const (
	B int64 = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
)

func ParseByte(s string) (int64, error) {
	if sl := len(s); sl > 0 {
		b := B
	BeginParse:
		switch sl--; s[sl] {
		case 'b', 'B':
			if sl == len(s)-1 {
				goto BeginParse
			}
		case 'k', 'K':
			b = KB
		case 'm', 'M':
			b = MB
		case 'g', 'G':
			b = GB
		case 't', 'T':
			b = TB
		case 'p', 'P':
			b = PB
		default:
			sl++
		}
		if sl == 0 {
			return 0, strconv.ErrSyntax
		}
		i, err := strconv.ParseInt(s[:sl], 10, 64)
		if err != nil {
			return 0, strconv.ErrSyntax
		}
		b *= i
		return b, nil
	}
	return 0, strconv.ErrSyntax
}
