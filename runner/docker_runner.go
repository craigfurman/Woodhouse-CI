package runner

import (
	"fmt"
	"io"
	"log"
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
	commandToRun := Chunk(job.Command)
	if len(commandToRun) == 0 {
		return fmt.Errorf("No arguments could be parsed from command: %s", job.Command)
	}

	args := []string{"run", "--rm", "busybox"}
	args = append(args, commandToRun...)
	containerCmd := exec.Command(r.DockerCmd, args...)

	containerCmd.Stdout = outputDest
	containerCmd.Stderr = outputDest
	if err := containerCmd.Start(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			close(status)
			return fmt.Errorf("running command: %s. Cause: %v", job.Command, err)
		}
	}

	go func() {
		if err := containerCmd.Wait(); err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				log.Printf("error waiting for job to finish: %v", err)
			}
		}

		// yep...
		status <- uint32(containerCmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())

		if err := outputDest.Close(); err != nil {
			log.Printf("error closing command output: %v", err)
		}
	}()

	return nil
}
