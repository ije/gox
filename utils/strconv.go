package utils

import (
	"errors"
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

// ParseBytes parses a bytes string
func ParseBytes(s string) (int64, error) {
	if p := len(s); p > 0 {
		b := B
	Start:
		switch p--; s[p] {
		case 'b', 'B':
			if p == len(s)-1 {
				goto Start
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
			p++
		}
		if p == 0 {
			return 0, strconv.ErrSyntax
		}
		i, err := strconv.ParseInt(s[:p], 10, 64)
		if err != nil {
			return 0, strconv.ErrSyntax
		}
		b *= i
		return b, nil
	}
	return 0, strconv.ErrSyntax
}

var errNaN = errors.New("NaN")

// ToNumber covert a 'number' interface to float64.
func ToNumber(v interface{}) (f float64, err error) {
	switch i := v.(type) {
	case string:
		f, err = strconv.ParseFloat(i, 64)
	case int:
		f = float64(i)
	case int8:
		f = float64(i)
	case int16:
		f = float64(i)
	case int32:
		f = float64(i)
	case int64:
		f = float64(i)
	case uint:
		f = float64(i)
	case uint8:
		f = float64(i)
	case uint16:
		f = float64(i)
	case uint32:
		f = float64(i)
	case uint64:
		f = float64(i)
	case float32:
		f = float64(i)
	case float64:
		f = i
	default:
		err = errNaN
	}
	return
}
