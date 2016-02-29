package valid

import (
	"testing"
)

func TestValid(t *testing.T) {
	check := func(ret, expected bool) {
		if ret != expected {
			t.Fatal("failed")
		}
	}

	check(IsNumber("12345"), true)
	check(IsNumber("12345", 0), true)
	check(IsNumber("12345", 4), false)
	check(IsNumber("12345", 5), true)
	check(IsNumber("12345", 6), false)
	check(IsNumber("1234", 5, 7), false)
	check(IsNumber("12345", 5, 7), true)
	check(IsNumber("12345", 7, 5), true)
	check(IsNumber("123456", 5, 7), true)
	check(IsNumber("123456", 7, 5), true)
	check(IsNumber("1234567", 5, 7), true)
	check(IsNumber("1234567", 7, 5), true)
	check(IsNumber("12345678", 5, 7), false)
	check(IsNumber("12345abc"), false)
	check(IsHexString("12345abc"), true)
	check(IsHexString("12345abcDEF"), true)
	check(IsHexString("12345abcDEFXYZ"), false)
	check(IsSlug("1234", 0), true)
	check(IsSlug("hello", 4), false)
	check(IsIPv4("192.168.1.1"), true)
	check(IsIPv4("127.0.0.1"), true)
	check(IsDomain("google.com"), true)
	check(IsDomain("i-je.mail.google.com"), true)
	check(IsDomain("i.je"), true)
	check(IsDomain("i.je"), true)
	check(IsIETFLangTag("en"), true)
	check(IsIETFLangTag("en-US"), true)
	check(IsIETFLangTag("en-909"), true)
	check(IsEmail("hi@google.com"), true)
	check(IsEmail("hi-go@google.com"), true)
	check(IsEmail("hi.go@google.com"), true)
	check(IsEmail("hi_go@google.com"), true)
	check(IsEmail("hi+go@google.com"), false)
	check(IsEmail("10753328@qq.com"), true)
}
