package utils

import "strings"

// SplitByFirstByte splits string by a char from the beginning.
func SplitByFirstByte(s string, c byte) (string, string) {
	i := strings.IndexByte(s, c)
	if i == -1 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

// SplitByLastByte splits string by a char from the end.
func SplitByLastByte(s string, c byte) (string, string) {
	i := strings.LastIndexByte(s, c)
	if i == -1 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

// Join2 joins two strings with a separator.
// eg. Join2("react", '@', "19.0.0") -> "react@19.0.0"
func Join2(a string, sep byte, b string) string {
	al := len(a)
	buf := make([]byte, al+1+len(b))
	copy(buf, a)
	buf[al] = sep
	copy(buf[al+1:], b)
	return string(buf)
}
