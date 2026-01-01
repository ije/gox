package term

import (
	"os"
)

// ClearLineRight erase the current line from the cursor to the end of the line
func ClearLineRight() {
	os.Stdout.WriteString("\033[0K")
}

// ClearLineLeft erase the current line from the beginning of the line to the cursor
func ClearLineLeft() {
	os.Stdout.WriteString("\033[1K")
}

// ClearLine erase the current line.
func ClearLine() {
	os.Stdout.WriteString("\033[2K")
}

// ClearScreen erase the entire screen
func ClearScreen() {
	os.Stdout.WriteString("\033[2J")
}
