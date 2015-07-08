package valid

import (
	"testing"
)

func TestValid(t *testing.T) {
	t.Log(IsNumber("12345") == true)
	t.Log(IsNumber("12345", 0) == true)
	t.Log(IsNumber("12345", 4) == false)
	t.Log(IsNumber("12345", 5) == true)
	t.Log(IsNumber("12345", 6) == false)
	t.Log(IsNumber("1234", 5, 7) == false)
	t.Log(IsNumber("12345", 5, 7) == true)
	t.Log(IsNumber("12345", 7, 5) == true)
	t.Log(IsNumber("123456", 5, 7) == true)
	t.Log(IsNumber("123456", 7, 5) == true)
	t.Log(IsNumber("1234567", 5, 7) == true)
	t.Log(IsNumber("1234567", 7, 5) == true)
	t.Log(IsNumber("12345678", 5, 7) == false)
	t.Log(IsNumber("12345abc") == false)
	t.Log(IsHexString("12345abc") == true)
	t.Log(IsHexString("12345abcDEF") == true)
	t.Log(IsHexString("12345abcDEFXYZ") == false)
	t.Log(IsIP("192.168.1.1") == true)
	t.Log(IsDomain("google.com") == true)
	t.Log(IsDomain("i-je.mail.google.com") == true)
	t.Log(IsDomain("i.je") == true)
	t.Log(IsDomain("i.je") == true)
	t.Log(IsEmail("hi@google.com") == true)
	t.Log(IsEmail("hi-go@google.com") == true)
	t.Log(IsEmail("hi.go@google.com") == true)
	t.Log(IsEmail("hi_go@google.com") == true)
	t.Log(IsEmail("hi+go@google.com") == false)
	t.Log(IsEmail("10753328@qq.com") == true)
}
