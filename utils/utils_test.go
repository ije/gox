package utils

import (
	"path"
	"strings"
	"testing"
)

func TestHtmlToText(t *testing.T) {
	s1 := "<h1>Lorem Ipsum</h1><p>&copy;&nbsp;2014</p><p><b><i>Lorem</i></b> &amp; ipsum dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.<p>"
	t.Log(HtmlToText(s1, 64, true))
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
		"  /a/c/b/  ",
		"E:\\One\\Design\\Photos\\DSC_123.JPG",
	} {
		t.Logf("%s (%v)", p, PathClean(p) == path.Clean(strings.Replace(strings.ToLower(strings.TrimSpace(p)), "\\", "/", -1)))
	}
}
