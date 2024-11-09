package term

import "fmt"

func BoldOn() {
	fmt.Print("\033[1m")
}

func BoldOff() {
	fmt.Print("\033[22m")
}

func UnderlineOn() {
	fmt.Print("\033[4m")
}

func UnderlineOff() {
	fmt.Print("\033[24m")
}
