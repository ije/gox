package term

import (
	"os"
)

var (
	CR  = []byte{'\r'}
	EOL = []byte{'\n'}
)

// Raw is an interface for reading raw input.
type Raw interface {
	Next() byte
}

// Confirm asks the user for a yes or no answer.
func Confirm(raw Raw, prompt string) (value bool) {
	os.Stdout.WriteString(Cyan("? ") + prompt + " " + Dim("(y/N)"))

	defer func() {
		ClearLine()
		os.Stdout.Write(CR)
	}()

	for {
		key := raw.Next()
		switch key {
		case 3, 27: // Ctrl+C, Escape
			os.Stdout.Write(EOL)
			os.Stdout.WriteString(Dim("Aborted."))
			os.Stdout.Write(EOL)
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
	os.Stdout.WriteString(Cyan("? ") + prompt + " " + Dim(defaultValue))
	buf := make([]byte, 1024)
	bufN := 0

LOOP:
	for {
		key := raw.Next()
		switch key {
		case 3, 27: // Ctrl+C, Escape
			os.Stdout.Write(EOL)
			os.Stdout.WriteString(Dim("Aborted."))
			os.Stdout.Write(EOL)
			os.Exit(0)
		case 13, 32: // Enter, Space
			break LOOP
		case 127, 8: // Backspace, Ctrl+H
			if bufN > 0 {
				bufN--
				os.Stdout.Write([]byte{'\b', ' ', '\b'})
			} else {
				ClearLine()
				os.Stdout.Write(CR)
				os.Stdout.WriteString(Cyan("? ") + prompt + " ")
			}
		default:
			if (key >= 'a' && key <= 'z') || (key >= '0' && key <= '9') || key == '-' {
				if bufN == 0 {
					ClearLine()
					os.Stdout.Write(CR)
					os.Stdout.WriteString(Cyan("? ") + prompt + " ")
				}
				buf[bufN] = key
				bufN++
				os.Stdout.WriteString(string(key))
			}
		}
	}
	if bufN > 0 {
		value = string(buf[:bufN])
	} else {
		value = defaultValue
	}
	os.Stdout.Write(CR)
	os.Stdout.WriteString(Green("✔ ") + prompt + " " + Dim(value))
	os.Stdout.Write(EOL)
	return
}

// Select asks the user to select an item from a list.
func Select(raw Raw, prompt string, items []string) (selected string) {
	os.Stdout.WriteString(Cyan("? ") + prompt)
	os.Stdout.Write(EOL)

	defer func() {
		MoveCursorUp(len(items) + 1)
		os.Stdout.WriteString(Green("✔ ") + prompt + " " + Dim(selected))
		os.Stdout.Write(EOL)
		for range items {
			ClearLine()
			os.Stdout.Write(EOL)
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
			os.Stdout.WriteString(Dim("Aborted."))
			os.Stdout.Write(EOL)
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
	// todo: scroll the screen if the items are too many
	for i, name := range items {
		os.Stdout.Write(CR)
		if i == selected {
			os.Stdout.WriteString("> " + name)
		} else {
			os.Stdout.WriteString("  " + Dim(name))
		}
		os.Stdout.Write(EOL)
	}
}
