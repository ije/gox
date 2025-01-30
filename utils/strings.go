package utils

import "strings"

// Split2 splits string by a char from the beginning.
func Split2(s string, c byte) (string, string) {
	i := strings.IndexByte(s, c)
	if i == -1 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

// Split2Last splits string by a char from the end.
func Split2Last(s string, c byte) (string, string) {
	i := strings.LastIndexByte(s, c)
	if i == -1 {
		return s, ""
	}
	return s[:i], s[i+1:]
}
