package term

import (
	"os"
)

func BoldOn() {
	os.Stdout.WriteString("\033[1m")
}

func BoldOff() {
	os.Stdout.WriteString("\033[22m")
}

func UnderlineOn() {
	os.Stdout.WriteString("\033[4m")
}

func UnderlineOff() {
	os.Stdout.WriteString("\033[24m")
}
