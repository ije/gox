package term

import "fmt"

func ClearLine() {
	fmt.Print("\033[2K")
}

func ClearLineRight() {
	fmt.Print("\033[K")
}

func ClearScreen() {
	fmt.Print("\033[2J")
}
