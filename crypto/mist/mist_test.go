package mist

import (
	"testing"
)

func TestMist(t *testing.T) {
	for l := 1; l <= 128; l++ {
		m := Hex.String(l)
		t.Log(m, len(m))
		m = Base64.String(l)
		t.Log(m, len(m))
	}
}
