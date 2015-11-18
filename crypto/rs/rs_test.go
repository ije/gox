package rs

import (
	"testing"
)

func TestDigital(t *testing.T) {
	for l := 1; l <= 64; l++ {
		rs := Digital.String(l)
		t.Log(rs, len(rs))
	}
}

func TestHex(t *testing.T) {
	for l := 1; l <= 64; l++ {
		rs := Hex.String(l)
		t.Log(rs, len(rs))
	}
}

func TestBase64(t *testing.T) {
	for l := 1; l <= 64; l++ {
		rs := Base64.String(l)
		t.Log(rs, len(rs))
	}
}
