package debug

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/ije/gox/term"
	"github.com/ije/gox/utils"
)

var (
	tempDir    string
	suPassword string
	processes  []*Process
	commands   map[string]func(args ...string) (ret string, err error)
	readlineEx *readline.Instance
)

var (
	Info = &term.ColorTerm{
		LinePrefix: "[debug] ",
		Color:      term.COLOR_GRAY,
	}
	Ok = &term.ColorTerm{
		LinePrefix: "[debug] ",
		Color:      term.COLOR_GREEN,
	}
	Warn = &term.ColorTerm{
		LinePrefix: "[debug] ",
		Color:      term.COLOR_RED,
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

	_, err = std.ColorPrintTo(pipe, color, string(p))
	if err == nil {
		n = len(p) // shoud return a right length
	}
	return
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

	_, err = std.ColorPrintTo(pipe, color, string(p))
	if err == nil {
		n = len(p) // shoud return a right length
	}
	return
}

func Run() {
	if len(processes) == 0 {
		return
	}

	AddCommand("restart", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if len(args) == 0 || utils.Contains(args, process.Name) {
				Info.Printf("Stopping %s...", process.Name)
				if err := process.Stop(); err != nil {
					Warn.Printf("Stop process %s failed: %v", process.Name, err)
				}
				Info.Printf("Starting %s...", process.Name)
				if err := process.Start(); err != nil {
					Warn.Printf("Start process %s failed: %v", process.Name, err)
					continue
				}
				Ok.Printf("The process %s has been restarted", process.Name)
			}
		}
		return
	})

	AddCommand("rebuild", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if len(args) == 0 || utils.Contains(args, process.Name) && (len(process.GoCode) > 0 || len(process.GoFile) > 0 || len(process.GoPkg) > 0) {
				Info.Printf("Stopping %s...", process.Name)
				err := process.Stop()
				if err != nil {
					Warn.Printf("Stop process %s failed: %v", process.Name, err)
				}
				Info.Printf("Building %s...", process.Name)
				err = process.Build()
				if err != nil {
					Warn.Printf("Build process %s failed: %v", process.Name, err)
					continue
				}
				Info.Printf("Starting %s...", process.Name)
				err = process.Start()
				if err != nil {
					Warn.Printf("Start process %s failed: %v", process.Name, err)
					continue
				}
				Ok.Printf("The process %s has been rebuild and restart", process.Name)
			}
		}
		return
	})

	AddCommand("exit|quit|bye", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if process.Stop() == nil {
				Ok.ColorPrintf(term.COLOR_NORMAL, "The process %s has been stoped", process.Name)
			}
		}
		return
	})

	AddCommand("status", func(args ...string) (ret string, err error) {
		now := time.Now()
		for _, process := range processes {
			if process.Process != nil {
				fmt.Fprintf(readlineEx.Stderr(), "%-30s %-10s pid %d, uptime %v\n", process.Name, strings.ToUpper(process.status), process.Pid, now.Sub(process.startTime))
			} else {
				fmt.Fprintf(readlineEx.Stderr(), "%-30s %-10s", process.Name, strings.ToUpper(process.status))
			}
		}
		return
	})

	var err error
	readlineEx, err = readline.NewEx(&readline.Config{
		Prompt:            "x$ ",
		HistoryFile:       path.Join(tempDir, "gox.debug/readline_history"),
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			if r == readline.CharCtrlZ {
				return r, false
			}
			return r, true
		},
	})
	if err != nil {
		panic(err)
	}
	defer readlineEx.Close()

	Ok.Pipe = readlineEx.Stdout()
	Info.Pipe = readlineEx.Stderr()
	Warn.Pipe = readlineEx.Stderr()

	var needSuPassword bool
	for _, process := range processes {
		if process.Sudo {
			needSuPassword = true
			break
		}
	}
	if needSuPassword {
		var password string
		for i := 0; i < 3; i++ {
			pw, err := readlineEx.ReadPassword("please enter the root password: ")
			if err != nil {
				panic(err)
			}

			output, err := exec.Command("/bin/bash", "-c", fmt.Sprintf(`echo "%s" | sudo -S -p "" -k whoami`, string(pw))).Output()
			if err != nil || strings.TrimSpace(string(output)) != "root" {
				fmt.Println("invalid root password")
				continue
			}

			password = string(pw)
			break
		}

		if len(password) == 0 {
			return
		}

		suPassword = password
	}

	for _, process := range processes {
		err := process.watch()
		if err != nil {
			Warn.Printf("watch process '%s' failed: %v", process.Name, err)
			continue
		}

		err = process.Build()
		if err != nil {
			Warn.Printf("Build process '%s' failed: %v", process.Name, err)
			continue
		}

		err = process.Start()
		if err != nil {
			Warn.Printf("Start process '%s' failed: %v", process.Name, err)
			continue
		}

		Ok.Printf("The process %s has been watched", process.Name)
	}

	for {
		line, err := readlineEx.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				for _, process := range processes {
					Info.Printf("Stopping %s...", process.Name)
					if err := process.Stop(); err != nil {
						Warn.Printf("Stop process %s failed: %v", process.Name, err)
					}
				}
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		ls := strings.Split(strings.TrimSpace(line), " ")
		cmd, args := ls[0], ls[1:]
		if handler, ok := commands[cmd]; ok {
			if ret, err := handler(args...); err != nil {
				Warn.Printf("%s: v", cmd, err)
			} else if len(ret) > 0 {
				Ok.Print(ret)
			}
			if cmd == "exit" || cmd == "quit" || cmd == "bye" {
				break
			}
		} else {
			Warn.Printf("Unknown command: %s", cmd)
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
