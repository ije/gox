package debug

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ije/gox/term"
	"github.com/ije/gox/utils"
)

type Process struct {
	Sudo             bool
	Name             string
	Status           string
	Code             string
	Path             string
	Args             []string
	LinkedPkgs       []string
	LinkedFiles      []string
	TermLinePrefix   string
	TermColorManager func(b []byte) term.Color
	watchingFiles    map[string]time.Time
	*os.Process
}

func (process *Process) Build() (err error) {
	process.Status = "building"
	defer func() {
		process.Status = ""
	}()

	for _, pkg := range process.LinkedPkgs {
		output, err := exec.Command("go", "install", pkg).CombinedOutput()
		if err != nil {
			return fmt.Errorf("install pkg '%s' failed: %v", pkg, err)
		} else if len(output) > 0 {
			return fmt.Errorf("install pkg '%s' failed: %s", pkg, string(output))
		}
	}

	if len(process.Code) > 0 {
		var processName string
		if len(process.Path) > 0 {
			_, processName = utils.SplitByLastByte(process.Path, os.PathSeparator)
			if len(processName) == 0 {
				processName = process.Path
			}
		} else {
			processName = strings.ToLower(process.Name)
		}
		exePath := path.Join(tempDir, "gox.debug", processName)
		goFile := exePath + ".go"

		if err := ioutil.WriteFile(goFile, []byte(process.Code), 0644); err != nil {
			return fmt.Errorf("create the '%s.go' failed: %v", strings.ToLower(process.Name), err)
		}

		if build.Default.GOOS == "windows" {
			exePath += ".exe"
		}
		output, err := exec.Command("go", "build", "-o", exePath, goFile).CombinedOutput()
		if err != nil || len(output) > 0 {
			return fmt.Errorf("build process '%s' failed: %v: %s", process.Name, err, string(output))
		}
		process.Path = exePath
	}
	return
}

func (process *Process) Start() (err error) {
	if len(process.Path) == 0 {
		return fmt.Errorf("missing path")
	}

	var cmd *exec.Cmd
	if process.Sudo && build.Default.GOOS != "windows" {
		exec.Command("/bin/bash", "-c", "sudo echo 0 > /dev/null").Run()
		cmd = exec.Command("/bin/bash", "-c", strings.Join(append([]string{"sudo", process.Path}, process.Args...), " "))
	} else {
		cmd = exec.Command(process.Path, process.Args...)
	}

	termLinePrefix := process.TermLinePrefix
	if len(termLinePrefix) == 0 {
		termLinePrefix = fmt.Sprintf("[%s] ", process.Name)
	}
	cmd.Stderr = &stderr{process.TermColorManager, &term.ColorTerm{LinePrefix: termLinePrefix}}
	cmd.Stdout = &stdout{process.TermColorManager, &term.ColorTerm{LinePrefix: termLinePrefix}}

	runErr := make(chan error)
	go func() {
		process.Status = "running"
		defer func() {
			process.Status = "stop"
		}()

		runErr <- cmd.Run()
	}()

	select {
	case err = <-runErr:
		return
	case <-time.After(time.Second):
	}

	if process.Sudo && build.Default.GOOS != "windows" {
		var processName string
		if len(process.Path) > 0 {
			_, processName = utils.SplitByLastByte(process.Path, os.PathSeparator)
			if len(processName) == 0 {
				processName = process.Path
			}
		} else {
			processName = strings.ToLower(process.Name)
		}
		output, err := exec.Command("pgrep", strings.ToLower(processName)).Output()
		if err != nil || len(output) == 0 {
			return fmt.Errorf("find child process failed: %v", err)
		}
		process.Process = nil
		for _, ps := range utils.ToLines(string(output)) {
			if pid, err := strconv.Atoi(ps); err == nil && pid > cmd.Process.Pid {
				process.Process, err = os.FindProcess(pid)
				if err != nil {
					return fmt.Errorf("find child process failed: %v", err)
				}
				break
			}
		}
		if process.Process == nil {
			return fmt.Errorf("find child process failed: %v", err)
		}
	} else {
		process.Process = cmd.Process
	}
	return
}

func (process *Process) Stop() (err error) {
	if process.Status == "running" && process.Process != nil {
		if process.Sudo && build.Default.GOOS != "windows" {
			output, err := exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo kill %d", process.Pid)).CombinedOutput()
			if err == nil && len(output) > 0 {
				err = fmt.Errorf("stop process '%s' failed: %s", process.Name, string(output))
			}
		} else {
			err = process.Kill()
		}
		if err == nil {
			process.Process = nil
			process.Status = "stop"
		}
	}
	return
}

func (process *Process) Listen() (err error) {
	err = process.Stop()
	if err != nil {
		return
	}

	err = process.Build()
	if err != nil {
		return
	}

	err = process.Start()
	if err != nil {
		return
	}

	if len(process.LinkedPkgs) == 0 && len(process.LinkedFiles) == 0 {
		return
	}

	if process.watchingFiles == nil {
		process.watchingFiles = map[string]time.Time{}
	}

	for _, pkg := range process.LinkedPkgs {
		if len(pkg) > 0 {
			process.watchingFiles[path.Join(build.Default.GOPATH, "pkg", build.Default.GOOS+"_"+build.Default.GOARCH, pkg+".a")] = time.Time{}
		}
	}
	for _, file := range process.LinkedFiles {
		if len(file) > 0 {
			process.watchingFiles[file] = time.Time{}
		}
	}

	process.watch()
	return
}

func (process *Process) watch() {
	if len(process.watchingFiles) == 0 {
		return
	}

	for path, prevModtime := range process.watchingFiles {
		fi, err := os.Stat(path)
		if err != nil && os.IsExist(err) {
			Warn.Printf("watch file '%s' failed: %v", path, err)
			continue
		}

		modtime := fi.ModTime()
		if !modtime.Equal(prevModtime) {
			process.watchingFiles[path] = modtime

			err = process.Stop()
			if err != nil {
				Warn.Print("Stop process '%s' unsuccessfully: %v", process.Name, err)
				continue
			}

			err = process.Build()
			if err != nil {
				Warn.Print("Rebuild process '%s' unsuccessfully: %v", process.Name, err)
				continue
			}

			err = process.Start()
			if err != nil {
				Warn.Print("Restart process '%s' unsuccessfully: %v", process.Name, err)
				continue
			}
			Ok.Print("The process '" + process.Name + "' has been rebuild and restart")
		}
	}

	time.AfterFunc(time.Second, process.watch)
}

func AddProcess(process *Process) error {
	if process == nil {
		return fmt.Errorf("nil process")
	}

	if len(process.Name) == 0 {
		if len(process.Path) > 0 {
			_, process.Name = utils.SplitByLastByte(process.Path, os.PathSeparator)
			if len(process.Name) == 0 {
				process.Name = process.Path
			}
		} else {
			return fmt.Errorf("missing process name")
		}
	}

	if len(process.Code) == 0 && len(process.Path) == 0 {
		return fmt.Errorf("missing process path")
	}

	processes = append(processes, process)
	return nil
}
