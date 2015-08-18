package runner

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

//go:generate counterfeiter -o fake_vcs_fetcher/fake_vcs_fetcher.go . VcsFetcher
type VcsFetcher interface {
	Fetch(repository string, outputSink io.Writer) (string, error)
}

type DockerRunner struct {
	DockerCmd  string
	VcsFetcher VcsFetcher
}

func NewDockerRunner(vcsFetcher VcsFetcher) *DockerRunner {
	return &DockerRunner{DockerCmd: "docker", VcsFetcher: vcsFetcher}
}

func (r *DockerRunner) Run(job jobs.Job, outputDest io.WriteCloser, status chan<- uint32) error {
	commandToRun := Chunk(job.Command)
	if len(commandToRun) == 0 {
		return fmt.Errorf("No arguments could be parsed from command: %s", job.Command)
	}

	if job.DockerImage == "" {
		return errors.New("you need to specify a docker image when using DockerRunner")
	}

	go func() {
		defer func() {
			if err := outputDest.Close(); err != nil {
				log.Printf("error closing command output: %v", err)
			}
		}()

		args := []string{"run", "--rm"}

		if job.GitRepository != "" {
			checkoutDir, err := r.VcsFetcher.Fetch(job.GitRepository, outputDest)

			defer func() {
				if err := os.RemoveAll(checkoutDir); err != nil {
					log.Printf("error removing checkout dir: %s, cause %v\n", checkoutDir, err)
				}
			}()

			if err != nil {
				log.Printf("error fetching repository from vcs: cause: %v\n", err)
				status <- uint32(1)
				return
			}

			args = append(args, "-v", fmt.Sprintf("%s:/woodhouse-workspace", checkoutDir), "--workdir", "/woodhouse-workspace")
		}

		args = append(args, job.DockerImage)
		args = append(args, commandToRun...)
		containerCmd := exec.Command(r.DockerCmd, args...)
		containerCmd.Stdout = outputDest
		containerCmd.Stderr = outputDest

		if err := containerCmd.Run(); err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				log.Printf("error running job: %v", err)
				status <- uint32(1)
				return
			}
		}

		// yep...
		status <- uint32(containerCmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus())
	}()

	return nil
}
