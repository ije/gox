package term

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/ije/gox/utils"
)

type ColorTerm struct {
	LinePrefix string
	Color      Color
	Pipe       io.Writer
}

func (term *ColorTerm) Print(args ...interface{}) (n int, err error) {
	return term.PrintTo(term.Pipe, fmt.Sprint(args...))
}

func (term *ColorTerm) Printf(format string, args ...interface{}) (n int, err error) {
	return term.PrintTo(term.Pipe, fmt.Sprintf(format, args...))
}

func (term *ColorTerm) ColorPrint(color Color, args ...interface{}) (n int, err error) {
	return term.ColorPrintTo(term.Pipe, color, fmt.Sprint(args...))
}

func (term *ColorTerm) ColorPrintf(color Color, format string, args ...interface{}) (n int, err error) {
	return term.ColorPrintTo(term.Pipe, color, fmt.Sprintf(format, args...))
}

func (term *ColorTerm) PrintTo(pipe io.Writer, s string) (n int, err error) {
	buf := bytes.NewBuffer(nil)
	for _, line := range utils.ToLines(s) {
		if term.Color > 30 && term.Color < 38 {
			_, err = fmt.Fprintf(buf, "\x1b[0;%dm%s%s\x1b[0m\n", term.Color, term.LinePrefix, line)
		} else {
			_, err = fmt.Fprintf(buf, "%s%s\n", term.LinePrefix, line)
		}
		if err != nil {
			return
		}
	}

	if pipe == nil {
		pipe = os.Stderr
	}

	return pipe.Write(buf.Bytes())
}

func (term *ColorTerm) ColorPrintTo(pipe io.Writer, color Color, s string) (n int, err error) {
	buf := bytes.NewBuffer(nil)
	for _, line := range utils.ToLines(s) {
		if color > 30 && color < 38 {
			_, err = fmt.Fprintf(buf, "\x1b[0;%dm%s%s\x1b[0m\n", color, term.LinePrefix, line)
		} else {
			_, err = fmt.Fprintf(buf, "%s%s\n", term.LinePrefix, line)
		}
		if err != nil {
			break
		}
	}

	if pipe == nil {
		pipe = os.Stderr
	}

	return pipe.Write(buf.Bytes())
}
