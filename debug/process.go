package debug

import (
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ije/gox/utils"
)

type Process struct {
	Sudo             bool
	Name             string
	Path             string
	Args             []string
	BuildArgs        []string
	GoCode           string
	GoFile           string
	GoPkg            string
	LinkedPkgs       []string
	LinkedFiles      []string
	LinkedPlatforms  []string
	BeforeBuild      func(*Process) error
	TermLinePrefix   string
	TermColorManager func(b []byte) TermColor
	status           string
	startTime        time.Time
	watchingFiles    map[string]time.Time
	*os.Process
}

func (process *Process) ProcessName() (processName string) {
	if len(process.Path) > 0 {
		_, processName = utils.SplitByLastByte(process.Path, os.PathSeparator)
		if len(processName) == 0 {
			processName = process.Path
		}
		return
	}

	processName = strings.ToLower(process.Name)
	return
}

func (process *Process) Build() (err error) {
	if process.status == "running" {
		process.Stop()
	}

	process.status = "building"
	defer func() {
		process.status = "stop"
	}()

	if len(process.GoCode) > 0 || len(process.GoFile) > 0 || len(process.GoPkg) > 0 {
		exePath := path.Join(tempDir, "gox.debug", process.ProcessName())
		if build.Default.GOOS == "windows" {
			exePath += ".exe"
		}

		var goFile string
		if len(process.GoCode) > 0 {
			goFile = exePath + ".go"
			if err := ioutil.WriteFile(goFile, []byte(process.GoCode), 0644); err != nil {
				return fmt.Errorf("create the '%s.go' failed: %v", strings.ToLower(process.Name), err)
			}
		} else if len(process.GoFile) > 0 {
			goFile = process.GoFile
			fi, err := os.Stat(goFile)
			if err != nil && os.IsNotExist(err) {
				return err
			} else if !strings.HasSuffix(goFile, ".go") || (err == nil && fi.IsDir()) {
				return fmt.Errorf("the 'GoFile' is not a go file")
			}
		} else {
			fi, err := os.Stat(path.Join(build.Default.GOPATH, "src", process.GoPkg))
			if err != nil && os.IsNotExist(err) {
				return err
			} else if err == nil && !fi.IsDir() {
				return fmt.Errorf("the 'GoPkg' is not a directory")
			}

			goFile = process.GoPkg
		}

		if process.BeforeBuild != nil {
			err = process.BeforeBuild(process)
			if err != nil {
				return fmt.Errorf("call BeforeBuild: %v", err)
			}
		}

		execArgs := []string{"build"}
		if len(process.BuildArgs) > 0 {
			execArgs = append(execArgs, process.BuildArgs...)
		}
		execArgs = append(execArgs, "-o", exePath, goFile)
		output, err := exec.Command("go", execArgs...).CombinedOutput()
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
		cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf(`echo "%s" | sudo -S -p "" %s %s`, suPassword, process.Path, strings.Join(process.Args, " ")))
	} else {
		cmd = exec.Command(process.Path, process.Args...)
	}

	termLinePrefix := process.TermLinePrefix
	if len(termLinePrefix) == 0 {
		termLinePrefix = fmt.Sprintf("[%s] ", process.Name)
	}

	var stdout io.Writer = os.Stdout
	var stderr io.Writer = os.Stderr
	if readlineEx != nil {
		stdout = readlineEx.Stdout()
		stderr = readlineEx.Stderr()
	}
	cmd.Stdout = &Stdio{process.TermColorManager, &Fmt{LinePrefix: termLinePrefix, Pipe: stdout}}
	cmd.Stderr = &Stdio{process.TermColorManager, &Fmt{LinePrefix: termLinePrefix, Pipe: stderr}}

	runErr := make(chan error)
	go func() {
		process.status = "running"
		runErr <- cmd.Run()
		process.status = "stop"
	}()

	select {
	case err = <-runErr:
		if err != nil {
			return
		}
	case <-time.After(time.Second):
	}

	if process.Sudo && build.Default.GOOS != "windows" {
		output, err := exec.Command("/bin/bash", "-c", fmt.Sprintf(`echo "%s" | sudo -S -p "" pgrep %s`, suPassword, process.ProcessName())).CombinedOutput()
		if err != nil || len(output) == 0 {
			return fmt.Errorf("find child process failed: %v", err)
		}
		for _, ps := range utils.ParseTextLines(string(output)) {
			if pid, err := strconv.Atoi(ps); err == nil {
				process.Process, err = os.FindProcess(pid)
				if err != nil {
					return fmt.Errorf("find child process failed: %v", err)
				}

				go func(pid int) {
					process.status = "running"
					for {
						time.Sleep(time.Second / 10)
						if _, err := os.FindProcess(pid); err != nil {
							break
						}
					}
					process.status = "stop"
				}(pid)

				break
			}
		}
	} else {
		process.Process = cmd.Process
	}

	process.startTime = time.Now().Add(-time.Second)
	return
}

func (process *Process) Stop() (err error) {
	if process.status == "running" && process.Process != nil {
		if process.Sudo && build.Default.GOOS != "windows" {
			output, err := exec.Command("/bin/bash", "-c", fmt.Sprintf(`echo "%s" | sudo -S -p "" kill -15 %d`, suPassword, process.Pid)).CombinedOutput()
			if err == nil && len(output) > 0 {
				err = fmt.Errorf("stop process '%s' failed: %s", process.Name, string(output))
			}
		} else {
			err = process.Signal(syscall.SIGTERM)
		}
		if err == nil {
			process.Process = nil
			process.status = "stop"
		}
	}
	return
}

func (process *Process) watch() (err error) {
	if len(process.watchingFiles) == 0 {
		process.watchingFiles = map[string]time.Time{}
	}

	currentPlatform := build.Default.GOOS + "_" + build.Default.GOARCH
	for _, pkg := range process.LinkedPkgs {
		output, err := exec.Command("go", "install", pkg).CombinedOutput()
		if err != nil {
			return fmt.Errorf("install pkg '%s' failed: %v", pkg, string(output))
		}

		if pkg = strings.TrimSpace(pkg); len(pkg) > 0 {
			process.watchingFiles[path.Join(build.Default.GOPATH, "pkg", currentPlatform, pkg+".a")] = time.Time{}
			if len(process.LinkedPlatforms) > 0 {
				for _, platform := range process.LinkedPlatforms {
					if platform = strings.TrimSpace(platform); len(platform) > 0 && platform != currentPlatform {
						process.watchingFiles[path.Join(build.Default.GOPATH, "pkg", platform, pkg+".a")] = time.Time{}
					}
				}
			}
		}
	}

	for _, file := range process.LinkedFiles {
		if file = strings.TrimSpace(file); len(file) > 0 {
			process.watchingFiles[file] = time.Time{}
		}
	}

	process.watchFileChange()
	return
}

func (process *Process) watchFileChange() {
	if len(process.watchingFiles) == 0 {
		return
	}

	defer time.AfterFunc(time.Second/3, process.watchFileChange)

	for path, prevModtime := range process.watchingFiles {
		fi, err := os.Stat(path)
		if err != nil {
			if os.IsExist(err) {
				Warn.Printf("watcher: process '%s': %v", process.Name, err)
			}
			continue
		}

		if modtime := fi.ModTime(); prevModtime.IsZero() {
			process.watchingFiles[path] = modtime
		} else if !modtime.Equal(prevModtime) {
			process.watchingFiles[path] = modtime
			process.Stop()

			err = process.Build()
			if err != nil {
				Warn.Printf("watcher: Build process '%s' failed: %v", process.Name, err)
				return
			}

			err = process.Start()
			if err != nil {
				Warn.Printf("watcher: Start process '%s' failed: %v", process.Name, err)
				return
			}

			if !prevModtime.IsZero() {
				Ok.Printf("watcher: The process '%s' has been rebuild and restart", process.Name)
			}
			return
		}
	}
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

	if len(process.GoCode) == 0 && len(process.GoFile) == 0 && len(process.GoPkg) == 0 && len(process.Path) == 0 {
		return fmt.Errorf("missing process path")
	}

	process.status = "stop"
	processes = append(processes, process)
	return nil
}
