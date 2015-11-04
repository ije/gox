package utils

import (
	"path"
	"strings"
	"testing"
)

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
		"  /a/c/b/  ",
		"E:\\One\\Design\\Photos\\DSC_123.JPG",
	} {
		cp := PathClean(p, true)
		t.Logf("%s -> %s (%v)", p, cp, cp == path.Clean(strings.Replace(strings.ToLower(strings.TrimSpace(p)), "\\", "/", -1)))
	}
}
