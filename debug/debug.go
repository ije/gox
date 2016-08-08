package debug

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

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
	Ok = &term.ColorTerm{
		Color: term.COLOR_GREEN,
	}
	Warn = &term.ColorTerm{
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
				if err := process.Stop(); err != nil {
					Warn.Printf("Stop process %s failed: %v", process.Name, err)
					continue
				}
				if err := process.Start(); err != nil {
					Warn.Printf("Restart process %s failed: %v", process.Name, err)
					continue
				}
				Ok.Printf("The process %s has been restarted", process.Name)
			}
		}
		return
	})

	AddCommand("rebuild", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if (len(args) == 0 || utils.Contains(args, process.Name)) && len(process.GoCode) > 0 {
				if err := process.Stop(); err != nil {
					Warn.Printf("Stop process %s failed: %v", process.Name, err)
					continue
				}
				if err := process.Build(); err != nil {
					Warn.Printf("Rebuild process %s failed: %v", process.Name, err)
					continue
				}
				if err := process.Start(); err != nil {
					Warn.Printf("Restart process %s failed: %v", process.Name, err)
					continue
				}
				Ok.Printf("The process %s has been rebuild and restart", process.Name)
			}
		}
		return
	})

	AddCommand("exit|bye|quit", func(args ...string) (ret string, err error) {
		for _, process := range processes {
			if process.Stop() == nil {
				Ok.ColorPrintf(term.COLOR_NORMAL, "The process %s has been stoped", process.Name)
			}
		}
		return
	})

	needSuPassword := false
	for _, process := range processes {
		if process.Sudo {
			needSuPassword = true
			break
		}
	}

	if needSuPassword {
		cl := term.NewCMDLine(nil)
		cl.AddStep("Please enter the SU Password:", func(input string) interface{} {
			output, err := exec.Command("/bin/bash", "-c", fmt.Sprintf(`echo "%s" | sudo -S -p "" -k whoami`, input)).Output()
			if err != nil || strings.TrimSpace(string(output)) != "root" {
				return false
			}

			suPassword = input
			return true
		})
		cl.Scan()
	}

	for _, process := range processes {
		err := process.Listen()
		if err != nil {
			Warn.Printf("Listen process '%s' failed: %v", process.Name, err)
			continue
		}

		Ok.Printf("The process %s has been listened", process.Name)

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
	}

	var err error
	readlineEx, err = readline.NewEx(&readline.Config{
		Prompt:      "x$ ",
		HistoryFile: path.Join(tempDir, "gox.debug/.rlh"),
	})
	if err != nil {
		panic(err)
	}
	defer readlineEx.Close()

	Ok.Pipe = readlineEx.Stdout()
	Warn.Pipe = readlineEx.Stderr()

	for {
		line, err := readlineEx.Readline()
		if err != nil {
			break
		}
		ls := strings.Split(line, " ")
		cmd, args := ls[0], ls[1:]
		if handler, ok := commands[cmd]; ok {
			if ret, err := handler(args...); err != nil {
				Warn.Printf("%s: v", cmd, err)
			} else if len(ret) > 0 {
				Ok.Print(ret)
			}
			if cmd == "bye" || cmd == "exit" || cmd == "quit" {
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

func UseHttpProxy(proxyRules map[string]string) (err error) {
	if len(proxyRules) == 0 {
		return
	}

	rules, err := json.Marshal(proxyRules)
	if err != nil {
		return
	}

	return AddProcess(&Process{
		Sudo:   true,
		Name:   "http-proxy",
		GoCode: fmt.Sprintf(HTTP_PROXY_SERVER_SRC, string(rules)),
	})
}

func init() {
	tempDir = os.TempDir()
	os.MkdirAll(path.Join(tempDir, "gox.debug"), 0755)
}
