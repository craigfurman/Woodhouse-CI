package runner

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type DockerRunner struct {
	DockerCmd string
}

func NewDockerRunner() *DockerRunner {
	return &DockerRunner{DockerCmd: "docker"}
}

func (r *DockerRunner) Run(job jobs.Job, outputDest io.WriteCloser, status chan<- uint32) error {
	defer outputDest.Close()

	args := []string{"run", "--rm", "busybox"}
	args = append(args, Chunk(job.Command)...)
	containerCmd := exec.Command(r.DockerCmd, args...)

	containerCmd.Stdout = outputDest
	containerCmd.Stderr = outputDest
	if err := containerCmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			close(status)
			return fmt.Errorf("running command: %s. Cause: %v", job.Command, err)
		}
	}

	// yep...
	status <- uint32(containerCmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())

	return nil
}
