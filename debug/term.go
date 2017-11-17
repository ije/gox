package debug

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ije/gox/os/term"
	"github.com/ije/gox/utils"
)

type ColorTerm struct {
	lock       sync.Mutex
	LinePrefix string
	Color      term.Color
	Pipe       io.Writer
}

func (term *ColorTerm) Print(args ...interface{}) (n int, err error) {
	msg := fmt.Sprintln(args...)
	if l := len(msg); l > 0 {
		msg = msg[:l-1]
	}
	return term.PrintTo(term.Pipe, msg)
}

func (term *ColorTerm) Printf(format string, args ...interface{}) (n int, err error) {
	return term.PrintTo(term.Pipe, fmt.Sprintf(format, args...))
}

func (term *ColorTerm) ColorPrint(color term.Color, args ...interface{}) (n int, err error) {
	msg := fmt.Sprintln(args...)
	if l := len(msg); l > 0 {
		msg = msg[:l-1]
	}
	return term.ColorPrintTo(term.Pipe, color, msg)
}

func (term *ColorTerm) ColorPrintf(color term.Color, format string, args ...interface{}) (n int, err error) {
	return term.ColorPrintTo(term.Pipe, color, fmt.Sprintf(format, args...))
}

func (term *ColorTerm) PrintTo(pipe io.Writer, s string) (n int, err error) {
	if pipe == nil {
		pipe = os.Stderr
	}

	term.lock.Lock()
	defer term.lock.Unlock()

	var i int
	for _, line := range utils.ParseLines(s, false) {
		if term.Color > 30 && term.Color < 38 {
			i, err = fmt.Fprintf(pipe, "\x1b[0;%dm%s%s\x1b[0m\n", term.Color, term.LinePrefix, line)
		} else {
			i, err = fmt.Fprintf(pipe, "%s%s\n", term.LinePrefix, line)
		}
		if err != nil {
			return
		}
		n += i
	}

	return
}

func (term *ColorTerm) ColorPrintTo(pipe io.Writer, color term.Color, s string) (n int, err error) {
	if pipe == nil {
		pipe = os.Stderr
	}

	term.lock.Lock()
	defer term.lock.Unlock()

	var i int
	for _, line := range utils.ParseLines(s, false) {
		if color > 30 && color < 38 {
			_, err = fmt.Fprintf(pipe, "\x1b[0;%dm%s%s\x1b[0m\n", color, term.LinePrefix, line)
		} else {
			_, err = fmt.Fprintf(pipe, "%s%s\n", term.LinePrefix, line)
		}
		if err != nil {
			return
		}
		n += i
	}

	return
}