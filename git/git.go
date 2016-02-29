package git

import (
	"errors"
	"os/exec"
	"strings"
	"time"
)

type Commit struct {
	Date    time.Time
	Hash    string
	Summary string
}

func LatestCommit(repoRoot string) (commit *Commit, err error) {
	cmd := exec.Command(
		"git",
		"log",
		"--pretty=format:%H|%ad|%s",
		"-1",
	)
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(output)
		if strings.Contains(msg, "bad default revision 'HEAD'") || strings.Contains(msg, "does not have any commits yet") {
			commit = &Commit{}
			err = nil
		} else {
			err = errors.New(msg)
		}
		return
	}

	info := strings.SplitN(string(output), "|", 3)
	date, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", info[1])
	if err != nil {
		return
	}
	commit = &Commit{
		Date:    date,
		Hash:    info[0],
		Summary: info[2],
	}
	return
}
