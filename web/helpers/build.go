package helpers

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type Build struct {
	JobId                string
	BuildNumber          string
	BytesAlreadyReceived int
	Name                 string
	Output               template.HTML
	ExitMessage          string
}

func PresentableJob(rj jobs.Build) Build {
	return Build{
		Name:        rj.Name,
		Output:      template.HTML(SanitisedHTML(string(rj.Output))),
		ExitMessage: message(rj),
	}
}

func SanitisedHTML(raw string) string {
	return strings.Replace(template.HTMLEscapeString(raw), "\n", "<br>", -1)
}

func message(build jobs.Build) string {
	if !build.Finished {
		return "Running"
	}
	if build.ExitStatus == 0 {
		return "Success"
	}
	return fmt.Sprintf("Failure: exit status %d", build.ExitStatus)
}
