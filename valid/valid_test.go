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
	t.Log(IsEmail("info@z.coffee") == true)
	t.Log(IsEmail("z-info@z.coffee") == true)
	t.Log(IsEmail("z.info@z.coffee") == true)
	t.Log(IsEmail("z_info@z.coffee") == true)
	t.Log(IsEmail("z+info@z.coffee") == false)
	t.Log(IsEmail("10753328@qq.com") == true)
}
