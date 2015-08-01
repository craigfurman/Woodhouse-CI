package helpers_test

import (
	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/web/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Build", func() {
	It("returns success when exit status is 0", func() {
		view := helpers.PresentableJob(jobs.Build{
			Job:        jobs.Job{Name: "ajob"},
			Output:     "output",
			ExitStatus: 0,
		})
		Expect(view).To(Equal(helpers.Build{
			Name:        "ajob",
			Output:      "output",
			ExitMessage: "Success",
		}))
	})

	It("returns failure when exit status is non-zero", func() {
		view := helpers.PresentableJob(jobs.Build{
			Job:        jobs.Job{Name: "ajob"},
			Output:     "output",
			ExitStatus: 42,
		})
		Expect(view).To(Equal(helpers.Build{
			Name:        "ajob",
			Output:      "output",
			ExitMessage: "Failure: exit status 42",
		}))
	})
})
