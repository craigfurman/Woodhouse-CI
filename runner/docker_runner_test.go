package runner_test

import (
	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DockerRunner", func() {
	var r *runner.DockerRunner

	BeforeEach(func() {
		r = &runner.DockerRunner{}
	})

	It("runs an arbitrary, fixed command", func() {
		job := jobs.Job{ID: "some-id", Name: "gob"}
		rj, err := r.Run(job)
		Expect(err).NotTo(HaveOccurred())
		Expect(rj).To(Equal(jobs.RunningJob{
			Job:    job,
			Output: "Hello world!\n",
		}))
	})

	PContext("when docker is not found in $PATH", func() {})
	PContext("when the command fails", func() {})
})
