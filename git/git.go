package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	errInvalidPrettyFormat = errors.New("invalid pretty format")
)

type Author struct {
	Name  string
	Email string
}

type Commit struct {
	Hash        string
	Date        time.Time
	Author      Author
	Title       string
	Description string
}

func Init(dir string, message string) (err error) {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, string(output))
		return
	}

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("%v: %s", err, string(output))
		return
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf(" %v: %s", err, string(output))
	}
	return
}

func InitUser(name string, email string) (err error) {
	output, _ := exec.Command("git", "config", "--global", "--get", "user.name").CombinedOutput()
	if name := strings.TrimSpace(string(output)); len(name) == 0 {
		err = exec.Command("git", "config", "--global", "user.name", name).Run()
		if err != nil {
			return
		}
	}

	output, _ = exec.Command("git", "config", "--global", "--get", "user.email").CombinedOutput()
	if email := strings.TrimSpace(string(output)); len(email) == 0 {
		err = exec.Command("git", "config", "--global", "user.email", email).Run()
		if err != nil {
			return
		}
	}

	return
}

func GetLatestCommit(dir string) (commit *Commit, err error) {
	cmd := exec.Command(
		"git",
		"log",
		"--pretty=format:%H\a\a\a%an\a\a\a%ae\a\a\a%at\a\a\a%s\a\a\a%b",
		"-1",
	)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New(string(output))
		return
	}

	commit, err = parseCommit(output)
	return
}

func GetLatestCommits(dir string, limit int) (commits []Commit, err error) {
	args := []string{
		"log",
		"--pretty=format:%H\a\a\a%an\a\a\a%ae\a\a\a%at\a\a\a%s\a\a\a%b\a\b\a",
		"-1000",
	}
	if limit > 0 {
		args[2] = fmt.Sprintf("-%d", limit)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.New(string(output))
		return
	}

	var commit *Commit
	for _, info := range bytes.Split(output, []byte{'\a', '\b', '\a'}) {
		if len(info) == 0 {
			continue
		}
		commit, err = parseCommit(info)
		if err != nil {
			return
		}
		commits = append(commits, *commit)
	}

	return
}

func parseCommit(data []byte) (commit *Commit, err error) {
	info := strings.SplitN(string(data), "\a\a\a", 6)
	if len(info) != 6 {
		err = errInvalidPrettyFormat
		return
	}

	ts, err := strconv.Atoi(info[3])
	if err != nil {
		err = errInvalidPrettyFormat
		return
	}

	commit = &Commit{
		Date: time.Unix(int64(ts), 0),
		Hash: info[0],
		Author: Author{
			Name:  info[1],
			Email: info[2],
		},
		Title:       info[4],
		Description: info[5],
	}
	return
}
