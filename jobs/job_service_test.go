package jobs_test

import (
	"errors"

	"github.com/craigfurman/woodhouse-ci/jobs"
	"github.com/craigfurman/woodhouse-ci/jobs/fake_job_repository"
	"github.com/craigfurman/woodhouse-ci/jobs/fake_job_runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service", func() {
	var (
		service *jobs.Service
		repo    *fake_job_repository.FakeRepository
		runner  *fake_job_runner.FakeRunner
	)

	BeforeEach(func() {
		repo = new(fake_job_repository.FakeRepository)
		runner = new(fake_job_runner.FakeRunner)
		service = &jobs.Service{
			Repository: repo,
			Runner:     runner,
		}
	})

	Describe("saving a job", func() {
		It("saves the job using the repository", func() {
			Expect(service.Save(&jobs.Job{Name: "freddo", Command: "whoami"})).To(Succeed())
			Expect(repo.SaveCallCount()).To(Equal(1))
			Expect(repo.SaveArgsForCall(0).Name).To(Equal("freddo"))
			Expect(repo.SaveArgsForCall(0).Command).To(Equal("whoami"))
		})

		Context("when saving fails", func() {
			BeforeEach(func() {
				repo.SaveReturns(errors.New("something went wrong"))
			})

			It("returns the error from the repository", func() {
				Expect(service.Save(nil)).To(MatchError("something went wrong"))
			})
		})
	})

	Describe("running a job", func() {
		BeforeEach(func() {
			job := jobs.Job{ID: "some-id", Name: "jerb", Command: "doStuff"}
			repo.FindByIdReturns(job, nil)
			runner.RunReturns(jobs.RunningJob{Job: job, Output: "some output", ExitStatus: 10}, nil)
		})

		It("runs the job and returns the result", func() {
			runningJob, err := service.RunJob("some-id")
			Expect(err).NotTo(HaveOccurred())
			Expect(runningJob).To(Equal(jobs.RunningJob{
				Job:        jobs.Job{ID: "some-id", Name: "jerb", Command: "doStuff"},
				Output:     "some output",
				ExitStatus: 10,
			}))

			Expect(runner.RunCallCount()).To(Equal(1))
		})

		Context("when the job cannot be found", func() {
			BeforeEach(func() {
				repo.FindByIdReturns(jobs.Job{}, errors.New("whoops!"))
			})

			It("returns error", func() {
				_, err := service.RunJob("bad-id")
				Expect(err).To(MatchError(ContainSubstring("running job with ID: bad-id")))
			})
		})

		Context("when running the cannot be started", func() {
			BeforeEach(func() {
				runner.RunReturns(jobs.RunningJob{}, errors.New("couldn't start job!"))
			})

			It("returns error", func() {
				_, err := service.RunJob("some-id")
				Expect(err).To(MatchError(ContainSubstring("starting job with ID: some-id")))
			})
		})
	})
})
