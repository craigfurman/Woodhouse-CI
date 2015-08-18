package vcs

import (
	"io"
	"io/ioutil"
	"os/exec"
)

type GitCloner struct{}

func (GitCloner) Fetch(repository string, outputSink io.Writer) (string, error) {
	tmpDir, err := ioutil.TempDir("", "woodhouse-git")
	if err != nil {
		return "", err
	}

	cloneCmd := exec.Command("git", "clone", "--recursive", repository, tmpDir)
	cloneCmd.Stdout = outputSink
	cloneCmd.Stderr = outputSink
	err = cloneCmd.Run()
	return tmpDir, err
}
