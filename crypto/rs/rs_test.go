package rs

import (
	"testing"
)

func TestDigital(t *testing.T) {
	for l := 1; l <= 512; l++ {
		rs := Digital.String(l)
		if (len(rs)) != l {
			t.Fatal("bad rs")
		}
	}
}

func TestHex(t *testing.T) {
	for l := 1; l <= 512; l++ {
		rs := Hex.String(l)
		if (len(rs)) != l {
			t.Fatal("bad rs")
		}
	}
}

func TestBase64(t *testing.T) {
	for l := 1; l <= 512; l++ {
		rs := Base64.String(l)
		if (len(rs)) != l {
			t.Fatal("bad rs")
		}
	}
}
