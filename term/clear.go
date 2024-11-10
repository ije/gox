package term

import (
	"os"
)

func ClearLine() {
	os.Stdout.WriteString("\033[2K")
}

func ClearScreen() {
	os.Stdout.WriteString("\033[2J")
}
