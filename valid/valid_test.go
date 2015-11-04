package valid

import (
	"testing"
)

func check(t *testing.T, ok bool) {
	if !ok {
		t.Fatal("failed")
	}
}

func TestValid(t *testing.T) {
	check(t, IsNumber("12345") == true)
	check(t, IsNumber("12345", 0) == true)
	check(t, IsNumber("12345", 4) == false)
	check(t, IsNumber("12345", 5) == true)
	check(t, IsNumber("12345", 6) == false)
	check(t, IsNumber("1234", 5, 7) == false)
	check(t, IsNumber("12345", 5, 7) == true)
	check(t, IsNumber("12345", 7, 5) == true)
	check(t, IsNumber("123456", 5, 7) == true)
	check(t, IsNumber("123456", 7, 5) == true)
	check(t, IsNumber("1234567", 5, 7) == true)
	check(t, IsNumber("1234567", 7, 5) == true)
	check(t, IsNumber("12345678", 5, 7) == false)
	check(t, IsNumber("12345abc") == false)
	check(t, IsHexString("12345abc") == true)
	check(t, IsHexString("12345abcDEF") == true)
	check(t, IsHexString("12345abcDEFXYZ") == false)
	check(t, IsSlug("1234", 0) == true)
	check(t, IsIP("192.168.1.1") == true)
	check(t, IsDomain("google.com") == true)
	check(t, IsDomain("i-je.mail.google.com") == true)
	check(t, IsDomain("i.je") == true)
	check(t, IsDomain("i.je") == true)
	check(t, IsIETFLangTag("en") == true)
	check(t, IsIETFLangTag("en-US") == true)
	check(t, IsIETFLangTag("en-909") == true)
	check(t, IsEmail("hi@google.com") == true)
	check(t, IsEmail("hi-go@google.com") == true)
	check(t, IsEmail("hi.go@google.com") == true)
	check(t, IsEmail("hi_go@google.com") == true)
	check(t, IsEmail("hi+go@google.com") == false)
	check(t, IsEmail("10753328@qq.com") == true)
}
