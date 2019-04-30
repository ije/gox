package debug

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ije/gox/utils"
)

type TermColor uint

const (
	T_COLOR_NORMAL TermColor = 0
	T_COLOR_RED    TermColor = 31
	T_COLOR_GREEN  TermColor = 32
	T_COLOR_YELLOW TermColor = 33
	T_COLOR_BLUE   TermColor = 34
	T_COLOR_PURPLE TermColor = 35
	T_COLOR_CYAN   TermColor = 36
	T_COLOR_GRAY   TermColor = 37
)

type Fmt struct {
	lock       sync.Mutex
	LinePrefix string
	Color      TermColor
	Pipe       io.Writer
}

func (f *Fmt) pipe() io.Writer {
	pipe := f.Pipe
	if pipe == nil {
		pipe = os.Stderr
	}
	return pipe
}

func (f *Fmt) Print(args ...interface{}) (n int, err error) {
	if f.Color > 0 {
		return f.ColorPrint(f.Color, args...)
	}
	return fmt.Fprint(f.pipe(), args...)
}

func (f *Fmt) Println(args ...interface{}) (n int, err error) {
	if f.Color > 0 {
		return f.ColorPrintln(f.Color, args...)
	}
	return fmt.Fprintln(f.pipe(), args...)
}

func (f *Fmt) Printf(format string, args ...interface{}) (n int, err error) {
	if f.Color > 0 {
		return f.ColorPrintf(f.Color, format, args...)
	}
	return fmt.Fprintf(f.pipe(), format, args...)
}

func (f *Fmt) ColorPrint(color TermColor, args ...interface{}) (n int, err error) {
	return f.colorFprint(f.Pipe, color, fmt.Sprint(args...))
}

func (f *Fmt) ColorPrintln(color TermColor, args ...interface{}) (n int, err error) {
	return f.colorFprint(f.Pipe, color, fmt.Sprintln(args...))
}

func (f *Fmt) ColorPrintf(color TermColor, format string, args ...interface{}) (n int, err error) {
	return f.colorFprint(f.Pipe, color, fmt.Sprintf(format, args...))
}

func (f *Fmt) colorFprint(pipe io.Writer, color TermColor, s string) (n int, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if pipe == nil {
		pipe = os.Stderr
	}

	for _, line := range utils.ParseTextLines(s) {
		var i int
		if color > 0 {
			i, err = fmt.Fprintf(pipe, "\x1b[0;%dm%s%s\x1b[0m\n", color, f.LinePrefix, line)
		} else {
			i, err = fmt.Fprintf(pipe, "%s%s\n", f.LinePrefix, line)
		}
		if err != nil {
			return
		}
		n += i
	}
	return
}

type Fmtx struct {
	parseColor func(b []byte) TermColor
	*Fmt
}

func (f *Fmtx) Write(p []byte) (n int, err error) {
	var color TermColor
	if f.parseColor != nil {
		color = f.parseColor(p)
	}

	n, err = f.ColorPrint(color, string(p))
	return
}
