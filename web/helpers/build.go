package helpers

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/craigfurman/woodhouse-ci/jobs"
)

func SanitisedHTML(raw []byte) template.HTML {
	colourizedText := ansiColours(raw)
	return template.HTML(strings.Replace(template.HTMLEscapeString(colourizedText), "\n", "<br>", -1))
}

func ansiColours(text []byte) string {
	// r := regexp.MustCompile(`{\\e\[(\d+)m(.*)}+`)
	r := regexp.MustCompile(`\\e\[(\d+)m(.*)`)
	matches := r.FindAll(text, -1)

	fmt.Println(string(bytes.Join(matches, []byte(":"))))
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

func Classes(build jobs.Build) string {
	if !build.Finished {
		return ""
	}

	if build.ExitStatus == 0 {
		return "passing"
	} else {
		return "failing"
	}
}
