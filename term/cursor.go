package term

import "fmt"

func MoveCursorToLineStart() {
	fmt.Print("\r")
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

func HideCursor() {
	fmt.Print("\033[?25l")
}

func ShowCursor() {
	fmt.Print("\033[?25h")
}
