package valid

import (
	"testing"
)

func TestValid(t *testing.T) {
	for _, test := range []struct {
		validator   func(string) bool
		testValue   string
		expectedRet bool
	}{
		{IsNumber, "123", true},
		{IsNumber, "123", true},
		{IsNumber, "123.456", true},
		{IsNumber, "123abc", false},
		{IsHexString, "123abc", true},
		{IsHexString, "123abcDEF", true},
		{IsHexString, "123abcDEFX", false},
		{IsSlug, "1234", true},
		{IsSlug, "hello", true},
		{IsSlug, "hello-world", true},
		{IsSlug, "hello world", false},
		{IsSlug, "hello.world", true},
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
		{IsEmail, "golang-2017@gmail.com", true},
		{IsEmail, "golang.2017@gmail.com", true},
		{IsEmail, "golang_2017@gmail.com", true},
		{IsEmail, "golang+2017@gmail.com", false},
		{IsEmail, "golang 2017@gmail.com", false},
		{IsEmail, "1234567890@qq.com", true},
	} {
		if test.expectedRet != test.validator(test.testValue) {
			t.Fatalf("no match: %s \n", test.testValue)
		}
	}
}
