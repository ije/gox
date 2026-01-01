package term

import (
	"fmt"
	"os"
)

// ANSI Escape Sequences: Cursor Movement
// @see: https://tldp.org/HOWTO/Bash-Prompt-HOWTO/x361.html

// MoveCursorTo moves the cursor to the specified position.
func MoveCursorTo(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

// MoveCursorUp moves the cursor up.
func MoveCursorUp(n int) {
	fmt.Printf("\033[%dA", n)
}

// MoveCursorDown moves the cursor down.
func MoveCursorDown(n int) {
	fmt.Printf("\033[%dB", n)
}

// MoveCursorRight moves the cursor right.
func MoveCursorRight(n int) {
	fmt.Printf("\033[%dC", n)
}

// MoveCursorLeft moves the cursor left.
func MoveCursorLeft(n int) {
	fmt.Printf("\033[%dD", n)
}

// SaveCursor saves the current cursor position (DEC).
func SaveCursor() {
	os.Stdout.WriteString("\0337")
}

// RestoreCursor restores the cursor position (DEC).
func RestoreCursor() {
	os.Stdout.WriteString("\0338")
}

// HideCursor hides the cursor.
func HideCursor() {
	os.Stdout.WriteString("\033[?25l")
}

// ShowCursor shows the cursor.
func ShowCursor() {
	os.Stdout.WriteString("\033[?25h")
}
