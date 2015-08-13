package helpers_test

import (
	"html/template"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/web/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Build", func() {
	It("returns success when exit status is 0", func() {
		b := jobs.Build{
			Job:        jobs.Job{Name: "ajob"},
			Output:     []byte("output"),
			ExitStatus: 0,
			Finished:   true,
		}
		view := helpers.PresentableJob(b)
		Expect(view).To(Equal(helpers.Build{
			Build:       b,
			Output:      "output",
			ExitMessage: "Success",
		}))
	})

	It("returns failure when exit status is non-zero", func() {
		b := jobs.Build{
			Job:        jobs.Job{Name: "ajob"},
			Output:     []byte("output"),
			ExitStatus: 42,
			Finished:   true,
		}
		view := helpers.PresentableJob(b)
		Expect(view).To(Equal(helpers.Build{
			Build:       b,
			Output:      template.HTML("output"),
			ExitMessage: "Failure: exit status 42",
		}))
	})

	It("returns pending when the build is not finished", func() {
		b := jobs.Build{
			Job:      jobs.Job{Name: "ajob"},
			Output:   []byte("output"),
			Finished: false,
		}
		view := helpers.PresentableJob(b)
		Expect(view).To(Equal(helpers.Build{
			Build:       b,
			Output:      template.HTML("output"),
			ExitMessage: "Running",
		}))
	})

	It("replaces newlines with html <br>", func() {
		view := helpers.PresentableJob(jobs.Build{
			Job:      jobs.Job{Name: "ajob"},
			Output:   []byte("some\nlines"),
			Finished: false,
		})
		Expect(view.Output).To(Equal(template.HTML("some<br>lines")))
	})

	It("still escapes all HTML in output", func() {
		view := helpers.PresentableJob(jobs.Build{
			Job:      jobs.Job{Name: "ajob"},
			Output:   []byte("<script>dangerous</script>"),
			Finished: false,
		})
		Expect(view.Output).To(Equal(template.HTML("&lt;script&gt;dangerous&lt;/script&gt;")))
	})
})
