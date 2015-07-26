package runner

import (
	"fmt"
	"os/exec"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type ArgChunker func(string) []string

//go:generate counterfeiter -o fake_command_runner/fake_command_runner.go . CommandRunner
type CommandRunner interface {
	CombinedOutput(cmd *exec.Cmd) ([]byte, uint32, error)
}

type DockerRunner struct {
	ArgChunker    ArgChunker
	CommandRunner CommandRunner
}

func (r *DockerRunner) Run(job jobs.Job) (jobs.RunningJob, error) {
	args := []string{"run", "--rm", "busybox"}
	args = append(args, r.ArgChunker(job.Command)...)
	containerCmd := exec.Command("docker", args...)
	output, exitStatus, err := r.CommandRunner.CombinedOutput(containerCmd)
	if err != nil {
		return jobs.RunningJob{}, fmt.Errorf("running command: %s. Cause: %v", job.Command, err)
	}

	return jobs.RunningJob{
		Job:        job,
		Output:     string(output),
		ExitStatus: exitStatus,
	}, nil
}
