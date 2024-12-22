package rand

import (
	"testing"
)

func TestRand(t *testing.T) {
	t.Log(Digital.String(4))
	t.Log(Hex.String(32))
	t.Log(Base64.String(64))
}
