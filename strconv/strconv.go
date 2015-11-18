package strconv

import (
	"strconv"
	"time"
)

const (
	B int64 = 1 << (10 * iota)
	KB
	MB
	GB
	TB
	PB
)

func ParseBytes(s string) (int64, error) {
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

func ParseDuration(s string) (time.Duration, error) {
	if sl := len(s); sl > 0 {
		t := time.Second
		withs := false
	BeginParse:
		switch sl--; s[sl] {
		case 's', 'S':
			if sl == len(s)-1 {
				withs = true
				goto BeginParse
			}
		case 'n', 'N':
			if withs {
				t = time.Nanosecond
			}
		case 'µ', 'u', 'U':
			if withs {
				t = time.Microsecond
			}
		case 'm', 'M':
			if withs {
				t = time.Millisecond
			} else {
				t = time.Minute
			}
		case 'h', 'H':
			t = time.Hour
		case 'd', 'D':
			t = 24 * time.Hour
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
		t *= time.Duration(i)
		return t, nil
	}
	return 0, strconv.ErrSyntax
}

func FormatInt(i int64, base int) string {
	return strconv.FormatInt(i, base)
}

func Itoa(i int) string {
	return strconv.Itoa(i)
}

func ParseInt(s string, base int, bitSize int) (int64, error) {
	return strconv.ParseInt(s, base, bitSize)
}

func ParseFloat(s string, bitSize int) (float64, error) {
	return strconv.ParseFloat(s, bitSize)
}

func Atoi(s string) (int, error) {
	return strconv.Atoi(s)
}
