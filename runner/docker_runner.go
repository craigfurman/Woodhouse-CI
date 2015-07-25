package runner

import (
	"os/exec"
	"strings"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type DockerRunner struct{}

func (DockerRunner) Run(job jobs.Job) (jobs.RunningJob, error) {
	args := []string{"run", "--rm", "busybox"}
	args = append(args, strings.Split(job.Command, " ")...)
	containerCmd := exec.Command("docker", args...)
	output, _ := containerCmd.CombinedOutput()
	return jobs.RunningJob{
		Job:    job,
		Output: string(output),
	}, nil
}
