package utils

import (
	"path"
	"strings"
	"testing"
)

func TestHtmlToText(t *testing.T) {
	html := "<div><h1>Lorem Ipsum</h1><p><b><i>lorem ipsum</i></b> dolor sit amet, consectetuer adipiscing elit, sed diam nonummy nibh euismod tincidunt ut laoreet dolore magna aliquam erat volutpat.<p><p>&copy;&nbsp;2014</p>&nbsp;&amp;&nbsp;lorem-ipsum<div>"
	text := HtmlToText(html, 160, true)
	t.Logf("%s (%d)", text, len(text))
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
		t.Logf("%s (%v)", p, PathClean(p, true) == path.Clean(strings.Replace(strings.ToLower(strings.TrimSpace(p)), "\\", "/", -1)))
	}
}
