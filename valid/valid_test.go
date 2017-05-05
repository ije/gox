package valid

import (
	"testing"
)

func TestValid(t *testing.T) {
	check := func(ret, expected bool) {
		if ret != expected {
			panic(nil)
		}
	}

	check(IsNumber("123"), true)
	check(IsNumber("123.456"), true)
	check(IsNumber("123abc"), false)
	check(IsHexString("123abc"), true)
	check(IsHexString("123abcDEF"), true)
	check(IsHexString("123abcDEF", 6), false)
	check(IsHexString("123abcDEF", 6, 12), true)
	check(IsHexString("123", 16, 32), false)
	check(IsHexString("123abcDEFXYZ"), false)
	check(IsSlug("1234", 0), true)
	check(IsSlug("hello", 4), false)
	check(IsIPv4("127.0.0.1"), true)
	check(IsIPv4("192.168.1.1"), true)
	check(IsIPv4("192.168.1.255"), true)
	check(IsIPv4("192.168.1.256"), false)
	check(IsDomain("google.com"), true)
	check(IsDomain("golang-google.com"), true)
	check(IsDomain("golang.google.com"), true)
	check(IsIETFLangTag("en"), true)
	check(IsIETFLangTag("en-US"), true)
	check(IsIETFLangTag("en-909"), true)
	check(IsEmail("go@gmail.com"), true)
	check(IsEmail("go-go@gmail.com"), true)
	check(IsEmail("go.go@gmail.com"), true)
	check(IsEmail("go_go@gmail.com"), true)
	check(IsEmail("go+go@gmail.com"), false)
	check(IsEmail("123456789@qq.com"), true)
}
