package utils

import (
	"path"
	"strings"
	"testing"
)

func TestParseLines(t *testing.T) {
	t.Log(strings.Join(ParseLines("abc\n123\rdef\r\r\n465\r\n\n\r\r\n789\n\r\n", true), "\\n"))
}

func TestPathClean(t *testing.T) {
	for _, p := range []string{
		"",
		".",
		"a.",
		"/",
		"./",
		"/.",
		"./.",
		"a/c",
		"a//c",
		"a/c/.",
		"a/c/b/..",
		"/../a/c",
		"/../a/b/../././/c",
		"/../a/../abc/123///ccc/../b/../././/c",
		"  /a/c/b/  ",
		"E:\\One\\Design\\Photos\\DSC_123.JPG",
	} {
		cp := CleanPath(p, true)
		cp2 := path.Clean(strings.Replace(strings.ToLower(strings.TrimSpace(p)), "\\", "/", -1))
		if cp != cp2 {
			t.Fatalf("%s -> %s (%v), should be %s", p, cp, cp == cp2, cp2)
		}
	}
}

func TestGetLocalIPs(t *testing.T) {
	t.Log(GetLocalIPs())
}
