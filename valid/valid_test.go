package valid

import (
	"testing"
)

func TestValid(t *testing.T) {
	for i, test := range []struct {
		validator func(string) bool
		value     string
		expected  bool
	}{
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
		{IsDomain, "google.com", true},
		{IsDomain, "golang-google.com", true},
		{IsDomain, "golang.google.com", true},
		{IsDomain, "golang_google.com", false},
		{IsIETFLangTag, "en", true},
		{IsIETFLangTag, "en-US", true},
		{IsIETFLangTag, "en-909", true},
		{IsIETFLangTag, "en909", false},
		{IsEmail, "golang@gmail.com", true},
		{IsEmail, "1234567890@gmail.com", true},
		{IsEmail, "go-lang@gmail.com", true},
		{IsEmail, "go.lang@gmail.com", true},
		{IsEmail, "go_lang@gmail.com", true},
		{IsEmail, "golang-@gmail.com", false},
		{IsEmail, "golang.@gmail.com", false},
		{IsEmail, "golang_@gmail.com", false},
		{IsEmail, "go+lang@gmail.com", false},
	} {
		if test.expected != test.validator(test.value) {
			t.Fatalf("no match(%d): %s", i, test.value)
		}
	}
}
