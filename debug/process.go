package debug

import (
	"encoding/json"
	"errors"
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
	Src              string
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

	if len(process.Src) > 0 {
		exePath := path.Join(tempDir, "gox.debug", strings.ToLower(process.Name))
		goFile := exePath + ".go"

		if err := ioutil.WriteFile(goFile, []byte(process.Src), 0644); err != nil {
			return fmt.Errorf("create the '%s.go' failed: %v", strings.ToLower(process.Name), err.Error())
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
		return errors.New("missing path")
	}

	var cmd *exec.Cmd
	if process.Sudo && build.Default.GOOS != "windows" {
		exec.Command("/bin/bash", "-c", "sudo echo 0 > /dev/null").Run()
		cmd = exec.Command("/bin/bash", "-c", strings.Join(append([]string{"sudo", process.Path}, process.Args...), " "))
	} else {
		cmd = exec.Command(process.Path, process.Args...)
	}

	cmd.Stderr = &stderr{process.TermColorManager, &term.ColorTerm{LinePrefix: process.TermLinePrefix}}
	cmd.Stdout = &stdout{process.TermColorManager, &term.ColorTerm{LinePrefix: process.TermLinePrefix}}

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
	case <-time.After(time.Second / 2):
	}

	if process.Sudo && build.Default.GOOS != "windows" {
		output, err := exec.Command("pgrep", strings.ToLower(process.Name)).Output()
		if err != nil || len(output) == 0 {
			return errors.New("find child process failed: " + err.Error())
		}
		process.Process = nil
		for _, pidstr := range utils.ToLines(string(output)) {
			if pid, err := strconv.Atoi(pidstr); err == nil && pid > cmd.Process.Pid {
				process.Process, err = os.FindProcess(pid)
				if err != nil {
					return errors.New("find child process failed: " + err.Error())
				}
				break
			}
		}
		if process.Process == nil {
			return errors.New("find child process failed: " + err.Error())
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
			errfmt.Printf("watch file '%s' failed: %v", path, err)
			continue
		}

		modtime := fi.ModTime()
		if !modtime.Equal(prevModtime) {
			process.watchingFiles[path] = modtime

			err = process.Stop()
			if err != nil {
				errfmt.Print("Stop process '" + process.Name + "' unsuccessfully: " + err.Error())
				continue
			}

			err = process.Build()
			if err != nil {
				errfmt.Print("Rebuild process '" + process.Name + "' unsuccessfully: " + err.Error())
				continue
			}

			err = process.Start()
			if err != nil {
				errfmt.Print("Restart process '" + process.Name + "' unsuccessfully: " + err.Error())
				continue
			}
			notice.Print("The process '" + process.Name + "' has been rebuild and restart")
		}
	}

	time.AfterFunc(time.Second, process.watch)
}

func AddProcess(process *Process) error {
	if process == nil {
		return errors.New("nil process")
	}

	if len(process.Name) == 0 {
		return errors.New("missing process name")
	}

	if len(process.Src) == 0 && len(process.Path) == 0 {
		return errors.New("missing process path")
	}

	processes = append(processes, process)
	return nil
}

func AddHttpProxyProcess(addr string, proxyRules map[string]string) (err error) {
	if len(addr) == 0 || len(proxyRules) == 0 {
		return
	}

	rulesJson, err := json.Marshal(proxyRules)
	if err != nil {
		return
	}

	return AddProcess(&Process{
		Sudo: true,
		Name: "gox.debug.http-proxy",
		Src:  fmt.Sprintf(HTTP_PROXY_SERVER_SRC, string(rulesJson), addr),
	})
}
