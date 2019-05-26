package valid

import (
	"regexp"
	"testing"
)

type test struct {
	validator func(string) bool
	value     string
	want      bool
}

func TestValid(t *testing.T) {
	for _, test := range []test{
		{IsNumber, "123", true},
		{IsNumber, "-123", true},
		{IsNumber, "123.456", true},
		{IsNumber, "123abc", false},
		{IsHexString, "123abc", true},
		{IsHexString, "123aBc", true},
		{IsHexString, "123aBcx", false},
		{IsSlug, "123", true},
		{IsSlug, "hello", true},
		{IsSlug, "hello-world", true},
		{IsSlug, "hello.world", true},
		{IsSlug, "helloworld-", false},
		{IsSlug, "helloworld.", false},
		{IsSlug, "hello world", false},
		{IsIPv4, "127.0.0.1", true},
		{IsIPv4, "192.168.1.1", true},
		{IsIPv4, "192.168.1.255", true},
		{IsIPv4, "192.168.1.256", false},
		{IsDomain, "golang.org", true},
		{IsDomain, "doc-golang.org", true},
		{IsDomain, "doc.golang.org", true},
		{IsDomain, "doc_golang.org", false},
		{IsIETFLangTag, "en", true},
		{IsIETFLangTag, "en-US", true},
		{IsIETFLangTag, "en-909", true},
		{IsIETFLangTag, "en909", false},
		{IsEmail, "doc@golang.org", true},
		{IsEmail, "123@golang.org", true},
		{IsEmail, "doc-dev@golang.org", true},
		{IsEmail, "doc.dev@golang.org", true},
		{IsEmail, "doc+dev@golang.org", false},
		{IsEmail, "doc_dev@golang.org", true},
		{IsEmail, "doc-@golang.org", false},
		{IsEmail, "doc.@golang.org", false},
		{IsEmail, "doc_@golang.org", false},
	} {
		if test.want != test.validator(test.value) {
			t.Fatalf("no match: %s", test.value)
		}
	}
}

var emails = map[string]bool{
	"doc@golang.org":     true,
	"123@golang.org":     true,
	"doc-dev@golang.org": true,
	"doc.dev@golang.org": true,
	"doc+dev@golang.org": true,
	"doc_dev@golang.org": true,
	"doc-@golang.org":    false,
	"doc.@golang.org":    false,
	"doc_@golang.org":    false,
	"doc+@golang.org":    false,
}

func BenchmarkIsEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for s, want := range emails {
			if want != IsEmail(s) {
				b.Fatalf("no match: %s", s)
			}
		}
	}
}

func BenchmarkIsEmailRegexp(b *testing.B) {
	b.StopTimer()
	emailRegexp := regexp.MustCompile(`^[a-zA-Z0-9]+((\.|\-|\_|\+)[a-zA-Z0-9]+)*@[a-zA-Z0-9]+((\.|\-)[a-zA-Z0-9]+)*\.[a-zA-Z]+$`)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		for s, want := range emails {
			if want != emailRegexp.MatchString(s) {
				b.Fatalf("no match: %s", s)
			}
		}
	}
}
