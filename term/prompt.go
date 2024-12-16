package term

import (
	"fmt"
	"os"
)

// Raw is an interface for reading raw input.
type Raw interface {
	Next() byte
}

// Confirm asks the user for a yes or no answer.
func Confirm(raw Raw, prompt string) (value bool) {
	fmt.Print(Cyan("? "))
	fmt.Print(prompt + " ")
	fmt.Print(Dim("(y/N)"))

	defer func() {
		ClearLine()
		fmt.Print("\r")
	}()

	for {
		key := raw.Next()
		switch key {
		case 3, 27: // Ctrl+C, Escape
			fmt.Print("\n")
			fmt.Print(Dim("Aborted."))
			fmt.Print("\n")
			os.Exit(0)
		case 13, 32: // Enter, Space
			return false
		case 'y':
			return true
		case 'n':
			return false
		}
	}
}

// Input asks the user for a string input.
func Input(raw Raw, prompt string, defaultValue string) (value string) {
	fmt.Print(Cyan("? "))
	fmt.Print(prompt + " ")
	fmt.Print(Dim(defaultValue))
	buf := make([]byte, 1024)
	bufN := 0

LOOP:
	for {
		key := raw.Next()
		switch key {
		case 3, 27: // Ctrl+C, Escape
			fmt.Print("\n")
			fmt.Print(Dim("Aborted."))
			fmt.Print("\n")
			os.Exit(0)
		case 13, 32: // Enter, Space
			break LOOP
		case 127, 8: // Backspace, Ctrl+H
			if bufN > 0 {
				bufN--
				fmt.Print("\b \b")
			} else {
				ClearLine()
				fmt.Print("\r")
				fmt.Print(Cyan("? "))
				fmt.Print(prompt + " ")
			}
		default:
			if (key >= 'a' && key <= 'z') || (key >= '0' && key <= '9') || key == '-' {
				if bufN == 0 {
					ClearLine()
					fmt.Print("\r")
					fmt.Print(Cyan("? "))
					fmt.Print(prompt + " ")
				}
				buf[bufN] = key
				bufN++
				fmt.Print(string(key))
			}
		}
	}
	if bufN > 0 {
		value = string(buf[:bufN])
	} else {
		value = defaultValue
	}
	fmt.Print("\r")
	fmt.Print(Green("✔ "))
	fmt.Print(prompt + " ")
	fmt.Print(Dim(value))
	fmt.Print("\n")
	return
}

// Select asks the user to select an item from a list.
func Select(raw Raw, prompt string, items []string) (selected string) {
	fmt.Print(Cyan("? "))
	fmt.Println(prompt)

	defer func() {
		MoveCursorUp(len(items) + 1)
		fmt.Print(Green("✔ "))
		fmt.Print(prompt + " ")
		fmt.Print(Dim(selected))
		fmt.Print("\n")
		for i := 0; i < len(items); i++ {
			ClearLine()
			fmt.Print("\n")
		}
		MoveCursorUp(len(items))
	}()

	HideCursor()
	defer ShowCursor()

	current := 0
	printSelectItems(items, current)

	for {
		key := raw.Next()
		switch key {
		case 3, 27: // Ctrl+C, Escape
			fmt.Print(Dim("Aborted."))
			fmt.Print("\n")
			ShowCursor()
			os.Exit(0)
		case 13, 32: // Enter, Space
			selected = items[current]
			return
		case 65, 16, 'p': // Up, ctrl+p, p
			if current > 0 {
				current--
				MoveCursorUp(len(items))
				printSelectItems(items, current)
			}
		case 66, 14, 'n': // Down, ctrl+n, n
			if current < len(items)-1 {
				current++
				MoveCursorUp(len(items))
				printSelectItems(items, current)
			}
		}
	}
}

func printSelectItems(items []string, selected int) {
	for i, name := range items {
		if i == selected {
			fmt.Println("\r> ", name)
		} else {
			fmt.Println("\r  ", Dim(name))
		}
	}
}
