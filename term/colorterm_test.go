package term

import (
	"testing"
)

func TestCP(t *testing.T) {
	cp := &ColorTerm{
		Color:      COLOR_GRAY,
		LinePrefix: "[gox] ",
	}
	cp.Print("hello world")
	cp.ColorPrint(COLOR_NORMAL, "hello world")
	cp.ColorPrint(COLOR_RED, "hello world")
	cp.ColorPrint(COLOR_GREEN, "hello world")
	cp.ColorPrint(COLOR_YELLOW, "hello world")
	cp.ColorPrint(COLOR_BLUE, "hello world")
	cp.ColorPrint(COLOR_PURPLE, "hello world")
	cp.ColorPrint(COLOR_CYAN, "hello world")
	cp.ColorPrint(COLOR_GRAY, "hello world")
	cp.ColorPrint(COLOR_RED, "hello world #1\nhello world #2")
}
