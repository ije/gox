package valid

import (
	"testing"
)

func TestValid(t *testing.T) {
	for i, test := range []struct {
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
		{IsSlug, "123", true},
		{IsSlug, "hello", true},
		{IsSlug, "hello-world", true},
		{IsSlug, "helloworld-", false},
		{IsSlug, "hello.world", false},
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
		{IsEmail, "golang-2@gmail.com", true},
		{IsEmail, "golang.2@gmail.com", true},
		{IsEmail, "golang_2@gmail.com", true},
		{IsEmail, "golang2_@gmail.com", false},
		{IsEmail, "golang+2@gmail.com", false},
		{IsEmail, "golang 2@gmail.com", false},
	} {
		if test.expectedRet != test.validator(test.testValue) {
			t.Fatalf("no match: %d %s \n", i, test.testValue)
		}
	}
}
