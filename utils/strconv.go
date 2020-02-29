package utils

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

// ParseBytes parses a bytes string
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

// ParseDuration parses a duration string
// TODO: parse format like '1d6h30m15s'?
func ParseDuration(s string) (time.Duration, error) {
	if sl := len(s); sl > 0 {
		t := time.Second
		endWithS := false
	BeginParse:
		switch sl--; s[sl] {
		case 's', 'S':
			if sl == len(s)-1 {
				endWithS = true
				goto BeginParse
			}
		case 'n', 'N':
			if endWithS {
				t = time.Nanosecond
			}
		case 'µ', 'u', 'U':
			if endWithS {
				t = time.Microsecond
			}
		case 'm', 'M':
			if endWithS {
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

		f, err := strconv.ParseFloat(s[:sl], 64)
		if err != nil {
			return 0, strconv.ErrSyntax
		}

		t = time.Duration(float64(t) * f)
		return t, nil
	}

	return 0, strconv.ErrSyntax
}
