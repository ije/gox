package term

import (
	"fmt"
	"os"
)

// ANSI Escape Sequences: Cursor Movement
// @see: https://tldp.org/HOWTO/Bash-Prompt-HOWTO/x361.html

func MoveCursorTo(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

func MoveCursorUp(n int) {
	fmt.Printf("\033[%dA", n)
}

func MoveCursorDown(n int) {
	fmt.Printf("\033[%dB", n)
}

func MoveCursorRight(n int) {
	fmt.Printf("\033[%dC", n)
}

func MoveCursorLeft(n int) {
	fmt.Printf("\033[%dD", n)
}

func SaveCursor() {
	os.Stdout.WriteString("\033[s")
}

func RestoreCursor() {
	os.Stdout.WriteString("\033[u")
}

func HideCursor() {
	os.Stdout.WriteString("\033[?25l")
}

func ShowCursor() {
	os.Stdout.WriteString("\033[?25h")
}
