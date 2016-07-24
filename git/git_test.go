package git

import (
	"testing"
)

func TestGitCommit(t *testing.T) {
	commit, err := GetLatestCommit("../")
	t.Log(commit, err)

	commits, err := GetLatestCommits("../", 0)
	t.Log(commits, err)
}
