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
