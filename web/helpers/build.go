package helpers

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

func SanitisedHTML(raw []byte) template.HTML {
	return template.HTML(strings.Replace(template.HTMLEscapeString(string(raw)), "\n", "<br>", -1))
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
