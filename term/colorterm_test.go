package term

import (
	"testing"
)

func TestCP(t *testing.T) {
	term := &ColorTerm{
		Color:      COLOR_GRAY,
		LinePrefix: "[gox] ",
	}
	term.Print("hello world")
	term.ColorPrint(COLOR_NORMAL, "hello world")
	term.ColorPrint(COLOR_RED, "hello world")
	term.ColorPrint(COLOR_GREEN, "hello world")
	term.ColorPrint(COLOR_YELLOW, "hello world")
	term.ColorPrint(COLOR_BLUE, "hello world")
	term.ColorPrint(COLOR_PURPLE, "hello world")
	term.ColorPrint(COLOR_CYAN, "hello world")
	term.ColorPrint(COLOR_GRAY, "hello world")
	term.ColorPrint(COLOR_RED, "hello world #1\nhello world #2")
}
