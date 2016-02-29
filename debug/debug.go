package debug

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/chzyer/readline"
	"github.com/ije/gox/term"
	"github.com/ije/gox/utils"
)

var (
	tempDir    string
	processes  []*Process
	commands   map[string]func(args ...string) (ret string, err error)
	readlineEx *readline.Instance
)

var (
	notice = &term.ColorTerm{
		Color: term.COLOR_GREEN,
	}
	errfmt = &term.ColorTerm{
		Color: term.COLOR_RED,
	}
)

type stdout struct {
	colorManager func(b []byte) term.Color
	*term.ColorTerm
}

func (std *stdout) Write(p []byte) (n int, err error) {
	var color term.Color
	var pipe io.Writer
	if std.colorManager != nil {
		color = std.colorManager(p)
	}
	if readlineEx != nil {
		pipe = readlineEx.Stdout()
	} else {
		pipe = os.Stdout
	}
	return std.ColorPrintTo(color, pipe, string(p))
}

type stderr struct {
	colorManager func(b []byte) term.Color
	*term.ColorTerm
}

func (std *stderr) Write(p []byte) (n int, err error) {
	var color term.Color
	var pipe io.Writer
	if std.colorManager != nil {
		color = std.colorManager(p)
	}
	if readlineEx != nil {
		pipe = readlineEx.Stderr()
	} else {
		pipe = os.Stderr
	}
	return std.ColorPrintTo(color, pipe, string(p))
}

func Run() {
	if len(processes) == 0 {
		return
	}

	AddCommand("restart", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if len(args) == 0 || utils.Contains(args, process.Name) {
				if err := process.Stop(); err != nil {
					errfmt.Print("Stop process '" + process.Name + "' unsuccessfully: " + err.Error())
					continue
				}
				if err := process.Start(); err != nil {
					errfmt.Print("Restart process '" + process.Name + "' unsuccessfully: " + err.Error())
					continue
				}
				notice.Print("The process '" + process.Name + "' has been restarted")
			}
		}
		return
	})

	AddCommand("rebuild", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if len(args) == 0 || utils.Contains(args, process.Name) && len(process.Src) > 0 {
				if err := process.Stop(); err != nil {
					errfmt.Print("Stop process '" + process.Name + "' unsuccessfully: " + err.Error())
					continue
				}
				if err := process.Build(); err != nil {
					errfmt.Print("Rebuild process '" + process.Name + "' unsuccessfully: " + err.Error())
					continue
				}
				if err := process.Start(); err != nil {
					errfmt.Print("Restart process '" + process.Name + "' unsuccessfully: " + err.Error())
					continue
				}
				notice.Print("The process '" + process.Name + "' has been rebuild and restart")
			}
		}
		return
	})

	AddCommand("exit|bye|quit", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if process.Stop() == nil {
				notice.ColorPrint(term.COLOR_GRAY, "The process '"+process.Name+"' has been stoped")
			}
		}

		return
	})

	var err error

	for _, process := range processes {
		if err := process.Listen(); err != nil {
			errfmt.Print("Listen process '" + process.Name + "' unsuccessfully: " + err.Error())
			continue
		}
		notice.Print("The process '" + process.Name + "' has been listened")
	}

	readlineEx, err = readline.NewEx(&readline.Config{
		Prompt:      "x$ ",
		HistoryFile: path.Join(tempDir, "gox.debug/.rlh"),
	})
	if err != nil {
		panic(err)
	}
	defer readlineEx.Close()

	notice.Pipe = readlineEx.Stdout()
	errfmt.Pipe = readlineEx.Stderr()

	for {
		line, err := readlineEx.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
		}
		ls := strings.Split(line, " ")
		cmd, args := ls[0], ls[:1]
		if handler, ok := commands[cmd]; ok {
			if ret, err := handler(args...); err != nil {
				errfmt.Print(cmd + ": " + err.Error())
			} else if len(ret) > 0 {
				notice.Print(ret)
			}
		}
		if cmd == "bye" || cmd == "exit" || cmd == "quit" {
			break
		}
	}
}

func AddCommand(names string, handler func(args ...string) (ret string, err error)) {
	if commands == nil {
		commands = map[string]func(args ...string) (ret string, err error){}
	}
	if len(names) > 0 && handler != nil {
		for _, name := range strings.Split(names, "|") {
			commands[name] = handler
		}
	}
}

func init() {
	tempDir = os.TempDir()
	os.MkdirAll(path.Join(tempDir, "gox.debug"), 0755)
}
