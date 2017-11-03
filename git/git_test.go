package git

import (
	"testing"
)

func TestGitCommit(t *testing.T) {
	ver, err := Version()
	t.Log(ver, err)

	commit, err := LatestCommit("../")
	t.Log(commit, err)

	commits, err := LatestCommits("../", 0)
	t.Log(commits, err)
}
