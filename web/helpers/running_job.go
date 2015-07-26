package helpers

import (
	"fmt"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type RunningJob struct {
	Name        string
	Output      string
	ExitMessage string
}

func PresentableJob(rj jobs.RunningJob) RunningJob {
	return RunningJob{
		Name:        rj.Name,
		Output:      rj.Output,
		ExitMessage: message(rj.ExitStatus),
	}
}

func message(status uint32) string {
	if status == 0 {
		return "Success"
	}
	return fmt.Sprintf("Failure: exit status %d", status)
}
