package strconv

import (
	"testing"
)

func TestParseDuration(t *testing.T) {
	t.Log(ParseDuration("12"))
	t.Log(ParseDuration("12s"))
	t.Log(ParseDuration("0.12s"))
	t.Log(ParseDuration("12ns"))
	t.Log(ParseDuration("12us"))
	t.Log(ParseDuration("12ms"))
	t.Log(ParseDuration("12m"))
	t.Log(ParseDuration("12h"))
	t.Log(ParseDuration("1d"))
	t.Log(ParseDuration("0.5d"))
	t.Log(ParseDuration("10d"))
	t.Log(ParseDuration("12x"))
	t.Log(ParseDuration("12w"))
	t.Log(ParseDuration("1d12h6m3s"))
}
