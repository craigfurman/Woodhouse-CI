package helpers

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

type Build struct {
	jobs.Build

	Output      template.HTML
	ExitMessage string

	BuildNumber          string
	BytesAlreadyReceived int
}

func PresentableJob(b jobs.Build) Build {
	return Build{
		Build:       b,
		Output:      template.HTML(SanitisedHTML(string(b.Output))),
		ExitMessage: Message(b),
	}
}

func SanitisedHTML(raw string) string {
	return strings.Replace(template.HTMLEscapeString(raw), "\n", "<br>", -1)
}

func Message(build jobs.Build) string {
	if !build.Finished {
		return "Running"
	}
	if build.ExitStatus == 0 {
		return "Success"
	}
	return fmt.Sprintf("Failure: exit status %d", build.ExitStatus)
}
