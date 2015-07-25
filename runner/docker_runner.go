package runner

import (
	"os/exec"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type DockerRunner struct{}

func (DockerRunner) Run(job jobs.Job) (jobs.RunningJob, error) {
	containerCmd := exec.Command("docker", "run", "--rm", "busybox", "echo", "Hello", "world!")
	output, _ := containerCmd.CombinedOutput()
	return jobs.RunningJob{
		Job:    job,
		Output: string(output),
	}, nil
}
